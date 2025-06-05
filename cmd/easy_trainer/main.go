package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"math"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"
)

// TongueTwister represents a single tongue twister with its metadata
type TongueTwister struct {
	Number   string `json:"number"`
	Date     string `json:"date"`
	Text     string `json:"text"`
	Stats    TwisterStats
	Score    float64
}

// TwisterStats holds statistical data about a tongue twister
type TwisterStats struct {
	WordCount      int
	CharCount      int
	VowelCount     int
	ConsonantCount int
	UniqueChars    int
	RepeatChars    int
	DifficultSounds     int    // Количество сложных звуков
	DifficultCombos     int    // Количество сложных сочетаний
	SoundComplexityScore float64 // Оценка сложности звуков
}

// Difficulty levels
const (
	Easy   = "Легкая"
	Medium = "Средняя"
	Hard   = "Сложная"
	Expert = "Очень сложная"
)

// Training modes
const (
	StandardMode   = "standard"
	TimedMode      = "timed"
	RepeatMode     = "repeat"
	ChallengeMode  = "challenge"
	PerfectionMode = "perfection" // New mode for perfection training
)

// DictionFocus represents areas to focus on for diction training
type DictionFocus struct {
	Name        string
	Description string
}

// Predefined diction focus areas
var dictionFocusAreas = []DictionFocus{
	{Name: "Артикуляция", Description: "Четкое произношение каждого звука"},
	{Name: "Ритм", Description: "Равномерный темп речи"},
	{Name: "Ударения", Description: "Правильное ударение в словах"},
	{Name: "Дыхание", Description: "Контроль дыхания при произношении"},
	{Name: "Скорость", Description: "Увеличение скорости без потери качества"},
}

// Сложные звуки и сочетания в русском языке
var (
	// Сложные звуки для произношения (шипящие, свистящие и др.)
	difficultSounds = []rune{'ж', 'ш', 'щ', 'ч', 'ц', 'р', 'л', 'ф', 'х'}
	
	// Сложные сочетания согласных
	difficultCombinations = []string{
		"ств", "здр", "вств", "стн", "нтг", "рдц", "стл", "нтск",
		"стск", "тск", "стр", "скр", "спр", "взр", "вдр", "встр",
		"всм", "рщ", "сч", "зщ", "жж", "жд", "жч", "шч", "щч",
		"чщ", "чт", "чш", "шт", "шц", "рл", "лр", "кр", "тр",
		"рт", "тч", "дж", "дз", "дц", "кс", "гз", "бз",
	}
	
	// Классификация звуков по группам для прогресса обучения
	soundProgressionGroups = []struct{
		Name string
		Sounds []rune
		Weight float64 // Вес сложности от 1 до 10
	}{
		{
			Name: "Простые гласные",
			Sounds: []rune{'а', 'о', 'у', 'э'},
			Weight: 1.0,
		},
		{
			Name: "Сложные гласные",
			Sounds: []rune{'ы', 'и', 'е', 'ё', 'ю', 'я'},
			Weight: 2.0,
		},
		{
			Name: "Простые согласные",
			Sounds: []rune{'м', 'н', 'п', 'б', 'т', 'д', 'к', 'г', 'в', 'ф'},
			Weight: 3.0,
		},
		{
			Name: "Свистящие",
			Sounds: []rune{'с', 'з', 'ц'},
			Weight: 5.0,
		},
		{
			Name: "Шипящие",
			Sounds: []rune{'ш', 'ж', 'щ', 'ч'},
			Weight: 7.0,
		},
		{
			Name: "Сонорные",
			Sounds: []rune{'р', 'л', 'й'},
			Weight: 8.0,
		},
	}
)

// UserPerformance хранит статистику выступления пользователя
type UserPerformance struct {
	SuccessRate      map[string]float64 // Успешность по типам звуков
	DifficultyRating map[string]float64 // Субъективная сложность категорий
	LastScores       []int              // Последние оценки для отслеживания прогресса
	TotalSessions    int                // Общее количество сессий
	AverageScore     float64            // Средний балл
}

// NewUserPerformance создает новый объект для отслеживания производительности
func NewUserPerformance() *UserPerformance {
	return &UserPerformance{
		SuccessRate:      make(map[string]float64),
		DifficultyRating: make(map[string]float64),
		LastScores:       make([]int, 0, 10),
		TotalSessions:    0,
		AverageScore:     3.0, // Начальное среднее значение
	}
}

func main() {
	// Parse command line flags
	jsonPathFlag := flag.String("json", "tongue_twisters/all_twisters.json", "Path to JSON file with tongue twisters")
	randomCountFlag := flag.Int("count", 5, "How many random tongue twisters to select for training")
	difficultyFlag := flag.String("difficulty", "all", "Difficulty level (easy, medium, hard, expert, all)")
	modeFlag := flag.String("mode", "standard", "Training mode (standard, timed, repeat, challenge, perfection)")
	timePerTwisterFlag := flag.Int("time", 30, "Seconds per tongue twister in timed mode")
	repetitionsFlag := flag.Int("reps", 3, "Number of repetitions in repeat mode")
	focusFlag := flag.Int("focus", 0, "Focus area for perfection mode (0-4, see documentation)")
	perfectionLevelFlag := flag.Int("level", 3, "Perfection level (1-5, higher is more demanding)")
	mixDifficultyFlag := flag.Bool("mix", true, "Mix different difficulty levels when selecting twisters")
	flag.Parse()

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Load and analyze tongue twisters
	twisters, err := loadTongueTwisters(*jsonPathFlag)
	if err != nil {
		fmt.Printf("Error loading tongue twisters: %v\n", err)
		os.Exit(1)
	}

	// Calculate statistics and score for each twister
	for i := range twisters {
		analyzeTwister(&twisters[i])
	}

	// Sort by difficulty score
	sort.Slice(twisters, func(i, j int) bool {
		return twisters[i].Score < twisters[j].Score
	})

	// Group by difficulty
	easyTwisters := filterTwistersByDifficulty(twisters, Easy)
	mediumTwisters := filterTwistersByDifficulty(twisters, Medium)
	hardTwisters := filterTwistersByDifficulty(twisters, Hard)
	expertTwisters := filterTwistersByDifficulty(twisters, Expert)

	// Print statistics
	fmt.Printf("Загружено %d скороговорок:\n", len(twisters))
	fmt.Printf("  %s: %d\n", Easy, len(easyTwisters))
	fmt.Printf("  %s: %d\n", Medium, len(mediumTwisters))
	fmt.Printf("  %s: %d\n", Hard, len(hardTwisters))
	fmt.Printf("  %s: %d\n", Expert, len(expertTwisters))
	fmt.Println()

	// Select twisters based on desired difficulty or mixed from all difficulties
	var trainingTwisters []TongueTwister
	
	if *mixDifficultyFlag && strings.ToLower(*difficultyFlag) == "all" {
		// Distribute the count among different difficulty levels
		totalCount := *randomCountFlag
		trainingTwisters = selectBalancedTwisters(easyTwisters, mediumTwisters, hardTwisters, expertTwisters, totalCount)
		fmt.Println("Выбраны скороговорки разной сложности для тренировки")
	} else {
		// Traditional selection based on single difficulty
		var selectedTwisters []TongueTwister
		switch strings.ToLower(*difficultyFlag) {
		case "easy":
			selectedTwisters = easyTwisters
		case "medium":
			selectedTwisters = mediumTwisters
		case "hard":
			selectedTwisters = hardTwisters
		case "expert":
			selectedTwisters = expertTwisters
		default:
			selectedTwisters = twisters
		}
	
		if len(selectedTwisters) == 0 {
			fmt.Println("Не найдено скороговорок выбранной сложности.")
			os.Exit(1)
		}
	
		// Select random twisters for training
		trainingTwisters = selectRandomTwisters(selectedTwisters, *randomCountFlag)
	}

	// Start the training session based on selected mode
	switch strings.ToLower(*modeFlag) {
	case TimedMode:
		runTimedTrainingSession(trainingTwisters, *timePerTwisterFlag)
	case RepeatMode:
		runRepeatTrainingSession(trainingTwisters, *repetitionsFlag)
	case ChallengeMode:
		runChallengeTrainingSession(trainingTwisters)
	case PerfectionMode:
		focusArea := *focusFlag
		if focusArea < 0 || focusArea >= len(dictionFocusAreas) {
			focusArea = 0
		}
		perfectionLevel := *perfectionLevelFlag
		if perfectionLevel < 1 || perfectionLevel > 5 {
			perfectionLevel = 3
		}
		runPerfectionTrainingSession(trainingTwisters, focusArea, perfectionLevel)
	default:
		runStandardTrainingSession(trainingTwisters)
	}
}

// loadTongueTwisters loads tongue twisters from a JSON file
func loadTongueTwisters(jsonPath string) ([]TongueTwister, error) {
	// Read the JSON file
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		// If the file doesn't exist at the specified path, try to find it in common locations
		if os.IsNotExist(err) {
			altPaths := []string{
				"all_twisters.json",
				"../tongue_twisters/all_twisters.json",
				"../../tongue_twisters/all_twisters.json",
			}

			for _, path := range altPaths {
				if data, err = os.ReadFile(path); err == nil {
					jsonPath = path
					break
				}
			}
		}
		
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", jsonPath, err)
		}
	}

	// Parse JSON
	var twisters []TongueTwister
	if err := json.Unmarshal(data, &twisters); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	fmt.Printf("Loaded tongue twisters from %s\n", jsonPath)
	return twisters, nil
}

// analyzeTwister calculates various statistics for a tongue twister and assigns a difficulty score
func analyzeTwister(twister *TongueTwister) {
	text := strings.ToLower(twister.Text)
	
	// Count words
	words := strings.Fields(text)
	twister.Stats.WordCount = len(words)
	
	// Count letters and classify them
	charMap := make(map[rune]int)
	
	for _, char := range text {
		if unicode.IsLetter(char) {
			twister.Stats.CharCount++
			charMap[char]++
			
			// Count vowels and consonants for Russian language
			if isRussianVowel(char) {
				twister.Stats.VowelCount++
			} else if unicode.IsLetter(char) {
				twister.Stats.ConsonantCount++
			}
			
			// Check if the character is a difficult sound
			if isRussianDifficultSound(char) {
				twister.Stats.DifficultSounds++
			}
		}
	}
	
	// Count unique and repeated characters
	twister.Stats.UniqueChars = len(charMap)
	for _, count := range charMap {
		if count > 1 {
			twister.Stats.RepeatChars += count - 1
		}
	}
	
	// Count difficult combinations
	twister.Stats.DifficultCombos = countDifficultCombinations(text)
	
	// Calculate sound complexity score
	twister.Stats.SoundComplexityScore = calculateSoundComplexity(text)
	
	// Calculate a difficulty score based on the statistics
	twister.Score = calculateDifficultyScore(twister.Stats)
}

// isRussianVowel checks if a character is a Russian vowel
func isRussianVowel(char rune) bool {
	vowels := []rune{'а', 'е', 'ё', 'и', 'о', 'у', 'ы', 'э', 'ю', 'я'}
	for _, v := range vowels {
		if char == v {
			return true
		}
	}
	return false
}

// isRussianDifficultSound checks if a character is considered difficult to pronounce
func isRussianDifficultSound(char rune) bool {
	for _, sound := range difficultSounds {
		if char == sound {
			return true
		}
	}
	return false
}

// countDifficultCombinations counts the number of difficult sound combinations in a text
func countDifficultCombinations(text string) int {
	count := 0
	for _, combo := range difficultCombinations {
		count += strings.Count(text, combo)
	}
	return count
}

// calculateSoundComplexity analyzes text for sound complexity based on progression groups
func calculateSoundComplexity(text string) float64 {
	text = strings.ToLower(text)
	
	// Calculate weighted presence of each sound group
	totalWeight := 0.0
	soundCount := 0
	
	for _, char := range text {
		if !unicode.IsLetter(char) {
			continue
		}
		
		soundCount++
		
		// Find weight for this sound
		for _, group := range soundProgressionGroups {
			for _, sound := range group.Sounds {
				if char == sound {
					totalWeight += group.Weight
					break
				}
			}
		}
	}
	
	// Avoid division by zero
	if soundCount == 0 {
		return 0
	}
	
	// Average complexity score weighted by frequency
	return totalWeight / float64(soundCount)
}

// calculateDifficultyScore assigns a numeric difficulty score to a tongue twister
func calculateDifficultyScore(stats TwisterStats) float64 {
	// Base difficulty is proportional to length
	score := float64(stats.WordCount) * 0.5
	
	// More characters increase difficulty
	score += float64(stats.CharCount) * 0.1
	
	// Consonant to vowel ratio affects difficulty
	consonantVowelRatio := 1.0
	if stats.VowelCount > 0 {
		consonantVowelRatio = float64(stats.ConsonantCount) / float64(stats.VowelCount)
	}
	score += consonantVowelRatio * 2.0
	
	// Repeated characters increase difficulty
	score += float64(stats.RepeatChars) * 0.3
	
	// Factor in difficult sounds and combinations
	score += float64(stats.DifficultSounds) * 0.5
	score += float64(stats.DifficultCombos) * 1.0
	
	// Include sound complexity score
	score += stats.SoundComplexityScore * 1.5
	
	return score
}

// getDifficultyLevel returns a human-readable difficulty level based on the score
func getDifficultyLevel(score float64) string {
	if score < 10 {
		return Easy
	} else if score < 20 {
		return Medium
	} else if score < 30 {
		return Hard
	} else {
		return Expert
	}
}

// filterTwistersByDifficulty returns tongue twisters of a specific difficulty level
func filterTwistersByDifficulty(twisters []TongueTwister, level string) []TongueTwister {
	var filtered []TongueTwister
	for _, twister := range twisters {
		if getDifficultyLevel(twister.Score) == level {
			filtered = append(filtered, twister)
		}
	}
	return filtered
}

// selectRandomTwisters selects n random tongue twisters from the given slice
func selectRandomTwisters(twisters []TongueTwister, n int) []TongueTwister {
	if n >= len(twisters) {
		return twisters
	}
	
	// Create a copy of the slice to avoid modifying the original
	shuffled := make([]TongueTwister, len(twisters))
	copy(shuffled, twisters)
	
	// Fisher-Yates shuffle
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	
	return shuffled[:n]
}

// runStandardTrainingSession conducts an interactive training session with the selected tongue twisters
func runStandardTrainingSession(twisters []TongueTwister) {
	fmt.Println("=== Начинаем стандартную тренировку ===")
	fmt.Printf("Выбрано %d скороговорок для практики.\n\n", len(twisters))
	
	for i, twister := range twisters {
		fmt.Printf("Скороговорка %d из %d:\n", i+1, len(twisters))
		fmt.Printf("Сложность: %s (%.1f)\n", getDifficultyLevel(twister.Score), twister.Score)
		fmt.Printf("Статистика: %d слов, %d букв (%d гласных, %d согласных)\n", 
			twister.Stats.WordCount, twister.Stats.CharCount, 
			twister.Stats.VowelCount, twister.Stats.ConsonantCount)
		fmt.Println()
		fmt.Println(twister.Text)
		fmt.Println()
		
		fmt.Println("Нажмите Enter для перехода к следующей скороговорке...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		fmt.Println(strings.Repeat("-", 60))
	}
	
	fmt.Println("=== Тренировка завершена ===")
}

// runTimedTrainingSession conducts a timed training session with the selected tongue twisters
func runTimedTrainingSession(twisters []TongueTwister, secondsPerTwister int) {
	fmt.Println("=== Начинаем тренировку на время ===")
	fmt.Printf("Выбрано %d скороговорок для практики. На каждую скороговорку %d секунд.\n\n", len(twisters), secondsPerTwister)
	
	for i, twister := range twisters {
		fmt.Printf("Скороговорка %d из %d:\n", i+1, len(twisters))
		fmt.Printf("Сложность: %s (%.1f)\n", getDifficultyLevel(twister.Score), twister.Score)
		fmt.Printf("Статистика: %d слов, %d букв (%d гласных, %d согласных)\n", 
			twister.Stats.WordCount, twister.Stats.CharCount, 
			twister.Stats.VowelCount, twister.Stats.ConsonantCount)
		fmt.Println()
		fmt.Println(twister.Text)
		fmt.Println()
		
		fmt.Printf("Время на практику: %d секунд. Нажмите Enter, когда будете готовы начать...\n", secondsPerTwister)
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		
		// Start timer
		fmt.Println("Время пошло! Повторяйте скороговорку...")
		
		// Create a channel for early completion
		done := make(chan bool)
		
		// Start a goroutine to listen for user input
		go func() {
			bufio.NewReader(os.Stdin).ReadBytes('\n')
			done <- true
		}()
		
		// Timer loop
		remaining := secondsPerTwister
		ticker := time.NewTicker(1 * time.Second)
		timerDone := false
		
		for !timerDone {
			select {
			case <-ticker.C:
				remaining--
				if remaining <= 0 {
					timerDone = true
				} else if remaining <= 5 {
					fmt.Printf("\rОсталось %d секунд...   ", remaining)
				}
			case <-done:
				timerDone = true
				fmt.Println("\rЗавершено раньше времени!                ")
			}
		}
		
		ticker.Stop()
		fmt.Println("\nВремя истекло!")
		fmt.Println(strings.Repeat("-", 60))
	}
	
	fmt.Println("=== Тренировка завершена ===")
}

// runRepeatTrainingSession conducts a training session with repeated practice of each tongue twister
func runRepeatTrainingSession(twisters []TongueTwister, repetitions int) {
	fmt.Println("=== Начинаем тренировку с повторениями ===")
	fmt.Printf("Выбрано %d скороговорок для практики. Каждую скороговорку нужно повторить %d раз.\n\n", 
		len(twisters), repetitions)
	
	for i, twister := range twisters {
		fmt.Printf("Скороговорка %d из %d:\n", i+1, len(twisters))
		fmt.Printf("Сложность: %s (%.1f)\n", getDifficultyLevel(twister.Score), twister.Score)
		fmt.Printf("Статистика: %d слов, %d букв (%d гласных, %d согласных)\n", 
			twister.Stats.WordCount, twister.Stats.CharCount, 
			twister.Stats.VowelCount, twister.Stats.ConsonantCount)
		fmt.Println()
		fmt.Println(twister.Text)
		fmt.Println()
		
		fmt.Println("Нажмите Enter, когда будете готовы начать повторения...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		
		for rep := 1; rep <= repetitions; rep++ {
			fmt.Printf("\rПовторение %d из %d. Нажмите Enter после прочтения...", rep, repetitions)
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		}
		
		fmt.Println("\nВы успешно повторили эту скороговорку!")
		fmt.Println(strings.Repeat("-", 60))
	}
	
	fmt.Println("=== Тренировка завершена ===")
}

// runChallengeTrainingSession conducts a challenging training session with increasing speed
func runChallengeTrainingSession(twisters []TongueTwister) {
	fmt.Println("=== Начинаем тренировку-вызов ===")
	fmt.Printf("Выбрано %d скороговорок для практики. Повторяйте каждую с увеличением скорости.\n\n", len(twisters))
	
	speeds := []string{"Медленно", "Средне", "Быстро", "Очень быстро"}
	
	for i, twister := range twisters {
		fmt.Printf("Скороговорка %d из %d:\n", i+1, len(twisters))
		fmt.Printf("Сложность: %s (%.1f)\n", getDifficultyLevel(twister.Score), twister.Score)
		fmt.Printf("Статистика: %d слов, %d букв (%d гласных, %d согласных)\n", 
			twister.Stats.WordCount, twister.Stats.CharCount, 
			twister.Stats.VowelCount, twister.Stats.ConsonantCount)
		fmt.Println()
		fmt.Println(twister.Text)
		fmt.Println()
		
		fmt.Println("Нажмите Enter, когда будете готовы начать испытание...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		
		for s, speed := range speeds {
			fmt.Printf("\rЧтение #%d: %s. Нажмите Enter после прочтения...", s+1, speed)
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		}
		
		fmt.Println("\nВы справились с вызовом!")
		fmt.Println(strings.Repeat("-", 60))
	}
	
	fmt.Println("=== Тренировка завершена ===")
}

// runPerfectionTrainingSession conducts a training session focused on perfecting diction and pronunciation
func runPerfectionTrainingSession(twisters []TongueTwister, focusArea int, perfectionLevel int) {
	focus := dictionFocusAreas[focusArea]
	
	fmt.Println("=== Начинаем тренировку идеальной дикции ===")
	fmt.Printf("Выбрано %d скороговорок для тренировки.\n", len(twisters))
	fmt.Printf("Фокус тренировки: %s - %s\n", focus.Name, focus.Description)
	fmt.Printf("Уровень требований: %d из 5\n\n", perfectionLevel)
	
	// Создаем профиль пользователя для этой сессии
	userProfile := NewUserPerformance()
	
	// Анализируем имеющиеся скороговорки для более умного выбора
	categorizedTwisters := categorizeTwistersForTraining(twisters, focusArea)
	
	// Display tips based on focus area
	fmt.Println("Рекомендации для тренировки:")
	switch focusArea {
	case 0: // Артикуляция
		fmt.Println("- Максимально чётко произносите каждую согласную")
		fmt.Println("- Следите за округлостью гласных звуков")
		fmt.Println("- Обратите внимание на положение языка и губ")
	case 1: // Ритм
		fmt.Println("- Следите за равномерностью произношения")
		fmt.Println("- Не торопитесь, выдерживайте одинаковый темп")
		fmt.Println("- Используйте метроном, если возможно (60-80 ударов в минуту)")
	case 2: // Ударения
		fmt.Println("- Выделяйте ударные слоги чуть сильнее")
		fmt.Println("- Не «проглатывайте» безударные гласные")
		fmt.Println("- Сохраняйте правильный ритмический рисунок слов")
	case 3: // Дыхание
		fmt.Println("- Сделайте глубокий вдох перед началом фразы")
		fmt.Println("- Распределите дыхание на всю фразу")
		fmt.Println("- Следите за контролем выдоха — он должен быть равномерным")
	case 4: // Скорость
		fmt.Println("- Начинайте медленно с идеальной артикуляцией")
		fmt.Println("- Постепенно увеличивайте скорость")
		fmt.Println("- При ускорении сохраняйте чёткость произношения")
	}
	fmt.Println()
	
	// Динамически определяем количество раундов в зависимости от уровня
	totalRounds := perfectionLevel + 2
	
	// Определяем прогрессию сложности
	difficulties := generateDifficultyProgression(perfectionLevel, totalRounds, 1.0)
	
	fmt.Println("Тренировка состоит из нескольких раундов с адаптивной сложностью")
	fmt.Println("Система будет подбирать скороговорки на основе вашего прогресса")
	fmt.Println()
	
	totalScore := 0
	
	for round := 1; round <= totalRounds; round++ {
		// Выбираем наиболее подходящую скороговорку для текущего раунда
		twister := selectOptimalTwister(categorizedTwisters, userProfile, round, totalRounds, focusArea)
		
		// Определяем текущую сложность
		currentDifficulty := difficulties[round-1]
		
		fmt.Printf("=== Раунд %d из %d (сложность %.1f) ===\n", round, totalRounds, currentDifficulty)
		fmt.Printf("Скороговорка: %s (%.1f)\n", getDifficultyLevel(twister.Score), twister.Score)
		
		// Выводим специфические особенности скороговорки в зависимости от фокуса тренировки
		presentTwisterFeatures(twister, focusArea)
		
		fmt.Println()
		fmt.Println(twister.Text)
		fmt.Println()
		
		// Даем конкретные советы по работе над этой скороговоркой
		provideFocusedAdvice(twister, focusArea, round, currentDifficulty)
		
		fmt.Println("\nНажмите Enter, когда будете готовы прочитать скороговорку...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		
		// Оценка производительности
		fmt.Print("Оцените свое произношение от 1 до 5: ")
		var score int
		fmt.Scanln(&score)
		if score < 1 {
			score = 1
		} else if score > 5 {
			score = 5
		}
		
		totalScore += score
		userProfile.LastScores = append(userProfile.LastScores, score)
		
		// Обновляем статистику пользователя
		updateUserPerformance(userProfile, twister, score, focusArea)
		
		// Адаптивно корректируем последующие раунды в зависимости от производительности
		if round < totalRounds {
			difficulties = adjustDifficulties(difficulties, round, score)
		}
		
		// Обратная связь и рекомендации
		provideFeedback(score, twister, focusArea)
		
		fmt.Println(strings.Repeat("-", 60))
	}
	
	// Анализ результатов сессии
	analyzeTrainingResults(userProfile, totalScore, totalRounds, focusArea)
}

// categorizeTwistersForTraining классифицирует скороговорки по специфическим характеристикам
func categorizeTwistersForTraining(twisters []TongueTwister, focusArea int) map[string][]TongueTwister {
	categories := make(map[string][]TongueTwister)
	
	// Базовые категории по сложности
	categories["easy"] = make([]TongueTwister, 0)
	categories["medium"] = make([]TongueTwister, 0)
	categories["hard"] = make([]TongueTwister, 0)
	categories["expert"] = make([]TongueTwister, 0)
	
	// Специализированные категории в зависимости от фокуса
	switch focusArea {
	case 0: // Артикуляция
		categories["шипящие"] = make([]TongueTwister, 0)
		categories["свистящие"] = make([]TongueTwister, 0)
		categories["сонорные"] = make([]TongueTwister, 0)
		categories["сложные_сочетания"] = make([]TongueTwister, 0)
	case 1: // Ритм
		categories["короткие"] = make([]TongueTwister, 0)
		categories["длинные"] = make([]TongueTwister, 0)
		categories["ритмичные"] = make([]TongueTwister, 0)
	case 3: // Дыхание
		categories["длинные_фразы"] = make([]TongueTwister, 0)
		categories["короткие_фразы"] = make([]TongueTwister, 0)
	case 4: // Скорость
		categories["повторяющиеся"] = make([]TongueTwister, 0)
		categories["скороговорки"] = make([]TongueTwister, 0)
	}
	
	// Классифицируем каждую скороговорку
	for _, twister := range twisters {
		// Классификация по уровню сложности
		diff := getDifficultyLevel(twister.Score)
		switch diff {
		case Easy:
			categories["easy"] = append(categories["easy"], twister)
		case Medium:
			categories["medium"] = append(categories["medium"], twister)
		case Hard:
			categories["hard"] = append(categories["hard"], twister)
		case Expert:
			categories["expert"] = append(categories["expert"], twister)
		}
		
		// Специализированные классификации
		switch focusArea {
		case 0: // Артикуляция
			text := strings.ToLower(twister.Text)
			if containsAny(text, []rune{'ш', 'щ', 'ж', 'ч'}) {
				categories["шипящие"] = append(categories["шипящие"], twister)
			}
			if containsAny(text, []rune{'с', 'з', 'ц'}) {
				categories["свистящие"] = append(categories["свистящие"], twister)
			}
			if containsAny(text, []rune{'р', 'л'}) {
				categories["сонорные"] = append(categories["сонорные"], twister)
			}
			if twister.Stats.DifficultCombos > 2 {
				categories["сложные_сочетания"] = append(categories["сложные_сочетания"], twister)
			}
		case 1: // Ритм
			if twister.Stats.WordCount <= 3 {
				categories["короткие"] = append(categories["короткие"], twister)
			} else if twister.Stats.WordCount >= 7 {
				categories["длинные"] = append(categories["длинные"], twister)
			}
			// Оценка ритмичности по повторам
			if twister.Stats.RepeatChars > twister.Stats.CharCount/3 {
				categories["ритмичные"] = append(categories["ритмичные"], twister)
			}
		case 3: // Дыхание
			if twister.Stats.CharCount > 60 {
				categories["длинные_фразы"] = append(categories["длинные_фразы"], twister)
			} else if twister.Stats.CharCount < 30 {
				categories["короткие_фразы"] = append(categories["короткие_фразы"], twister)
			}
		case 4: // Скорость
			if twister.Stats.RepeatChars > twister.Stats.CharCount/4 {
				categories["повторяющиеся"] = append(categories["повторяющиеся"], twister)
			}
			if strings.Contains(strings.ToLower(twister.Text), "скороговорк") {
				categories["скороговорки"] = append(categories["скороговорки"], twister)
			}
		}
	}
	
	return categories
}

// containsAny проверяет, содержит ли строка хотя бы один из указанных символов
func containsAny(s string, chars []rune) bool {
	for _, c := range chars {
		if strings.ContainsRune(s, c) {
			return true
		}
	}
	return false
}

// generateDifficultyProgression создает прогрессию сложности для тренировки
func generateDifficultyProgression(level, rounds int, startingDiff float64) []float64 {
	result := make([]float64, rounds)
	maxDiff := float64(level) * 1.5
	
	if maxDiff < startingDiff {
		maxDiff = startingDiff
	}
	
	step := (maxDiff - startingDiff) / float64(rounds-1)
	
	for i := 0; i < rounds; i++ {
		result[i] = startingDiff + float64(i)*step
		// Применяем небольшую случайность к каждому значению (±10%)
		randomFactor := 0.9 + rand.Float64()*0.2
		result[i] *= randomFactor
	}
	
	return result
}

// selectOptimalTwister выбирает оптимальную скороговорку для текущего этапа тренировки
func selectOptimalTwister(categories map[string][]TongueTwister, profile *UserPerformance, round, totalRounds, focusArea int) TongueTwister {
	// Определяем прогресс тренировки (от 0.0 до 1.0)
	progress := float64(round-1) / float64(totalRounds-1)
	
	// Выбираем категорию в зависимости от прогресса и фокуса
	var category string
	var candidateTwisters []TongueTwister
	
	// Если это первый раунд и есть легкие скороговорки, начинаем с них
	if round == 1 && len(categories["easy"]) > 0 {
		category = "easy"
	} else if round == totalRounds && len(categories["expert"]) > 0 {
		// Для последнего раунда выбираем сложные
		category = "expert"
	} else {
		// В зависимости от фокуса выбираем категории
		switch focusArea {
		case 0: // Артикуляция
			// В зависимости от прогресса меняем фокус
			if progress < 0.3 {
				category = randomChoice([]string{"easy", "medium", "свистящие"})
			} else if progress < 0.6 {
				category = randomChoice([]string{"medium", "шипящие", "свистящие"})
			} else {
				category = randomChoice([]string{"hard", "expert", "сонорные", "сложные_сочетания"})
			}
		case 1: // Ритм
			if progress < 0.4 {
				category = randomChoice([]string{"короткие", "easy", "medium"})
			} else if progress < 0.7 {
				category = randomChoice([]string{"medium", "ритмичные"})
			} else {
				category = randomChoice([]string{"длинные", "hard", "ритмичные"})
			}
		case 3: // Дыхание
			if progress < 0.4 {
				category = randomChoice([]string{"короткие_фразы", "easy"})
			} else {
				category = randomChoice([]string{"длинные_фразы", "medium", "hard"})
			}
		case 4: // Скорость
			if progress < 0.3 {
				category = "easy"
			} else if progress < 0.6 {
				category = randomChoice([]string{"medium", "повторяющиеся"})
			} else {
				category = randomChoice([]string{"hard", "expert", "повторяющиеся"})
			}
		default:
			// По умолчанию используем сложность в зависимости от прогресса
			if progress < 0.3 {
				category = "easy"
			} else if progress < 0.6 {
				category = "medium"
			} else if progress < 0.9 {
				category = "hard"
			} else {
				category = "expert"
			}
		}
	}
	
	// Проверяем наличие скороговорок в выбранной категории
	if twisters, ok := categories[category]; ok && len(twisters) > 0 {
		candidateTwisters = twisters
	} else {
		// Если категория пуста, берем скороговорки соответствующей сложности
		if progress < 0.3 {
			candidateTwisters = categories["easy"]
		} else if progress < 0.6 {
			candidateTwisters = categories["medium"]
		} else if progress < 0.9 {
			candidateTwisters = categories["hard"]
		} else {
			candidateTwisters = categories["expert"]
		}
		
		// Если и эта категория пуста, берем любую имеющуюся категорию
		if len(candidateTwisters) == 0 {
			for _, twisters := range categories {
				if len(twisters) > 0 {
					candidateTwisters = twisters
					break
				}
			}
		}
	}
	
	// Если категория существует, выбираем случайную скороговорку из нее
	if len(candidateTwisters) > 0 {
		// Получаем случайный индекс
		index := rand.Intn(len(candidateTwisters))
		return candidateTwisters[index]
	}
	
	// Запасной вариант - если нет подходящих скороговорок
	fmt.Println("ПРЕДУПРЕЖДЕНИЕ: Не найдено подходящих скороговорок, используется первая доступная")
	for _, twisters := range categories {
		if len(twisters) > 0 {
			return twisters[0]
		}
	}
	
	// Если вообще нет скороговорок (что странно, так как массив не должен быть пустым)
	panic("Ошибка: нет доступных скороговорок для тренировки")
}

// randomChoice выбирает случайный элемент из списка
func randomChoice(options []string) string {
	if len(options) == 0 {
		return ""
	}
	return options[rand.Intn(len(options))]
}

// presentTwisterFeatures отображает специфические особенности скороговорки
func presentTwisterFeatures(twister TongueTwister, focusArea int) {
	switch focusArea {
	case 0: // Артикуляция
		fmt.Printf("Сложные звуки: ")
		printComplexSounds(twister.Text)
		fmt.Printf("Сложные сочетания: %d\n", twister.Stats.DifficultCombos)
	case 1: // Ритм
		fmt.Printf("Ритмическая структура: ")
		printRhythmicStructure(twister.Text)
		fmt.Printf("Количество слогов: %d, Слов: %d\n", 
			countSyllables(twister.Text), twister.Stats.WordCount)
	case 2: // Ударения
		fmt.Printf("Обратите внимание на ударения в словах\n")
		highlightStressPatterns(twister.Text)
	case 3: // Дыхание
		fmt.Printf("Длина фразы: %d слов (%.1f букв на слово)\n", 
			twister.Stats.WordCount, float64(twister.Stats.CharCount)/float64(twister.Stats.WordCount))
		fmt.Printf("Общее количество символов: %d\n", twister.Stats.CharCount)
	case 4: // Скорость
		fmt.Printf("Сложность для скорости: %.1f\n", twister.Score)
		fmt.Printf("Сложные сочетания: %d, Повторения: %d\n", 
			twister.Stats.DifficultCombos, twister.Stats.RepeatChars)
		// Оценка примерного времени произношения
		fmt.Printf("Ориентировочное время произношения: %.1f сек\n", 
			float64(twister.Stats.CharCount)*0.1)
	}
}

// countSyllables подсчитывает количество слогов в тексте
func countSyllables(text string) int {
	words := strings.Fields(text)
	total := 0
	
	for _, word := range words {
		total += countRussianSyllables(word)
	}
	
	return total
}

// provideFocusedAdvice дает конкретные советы по работе над этой скороговоркой
func provideFocusedAdvice(twister TongueTwister, focusArea int, round int, difficulty float64) {
	fmt.Println("Фокус раунда:")
	
	switch focusArea {
	case 0: // Артикуляция
		suggestArticulationFocus(twister.Text, round)
		// Определяем наиболее сложные звуки в этой скороговорке
		highlightDifficultSounds(twister.Text)
	case 1: // Ритм
		suggestRhythmFocus(twister.Text, round)
		// Дополнительный совет по ритму
		if twister.Stats.WordCount > 5 {
			fmt.Println("Следите за равномерностью произношения длинной фразы")
		} else {
			fmt.Println("Сконцентрируйтесь на ровном ритме коротких слов")
		}
	case 2: // Ударения
		fmt.Printf("Правильные ударения в словах (уровень %.1f)\n", difficulty)
		// Указываем, на какие слова особенно обратить внимание
		highlightStressPatterns(twister.Text)
	case 3: // Дыхание
		suggestBreathingPattern(twister.Text, round)
		// Дополнительный совет по дыханию
		if twister.Stats.CharCount > 50 {
			fmt.Println("Делайте глубокий вдох перед началом этой длинной фразы")
		}
	case 4: // Скорость
		suggestSpeedFocus(twister.Stats.WordCount, round, round+2)
		// Дополнительный совет по скорости
		if twister.Stats.DifficultCombos > 2 {
			fmt.Println("Особое внимание уделите сложным сочетаниям звуков")
		}
	}
}

// highlightDifficultSounds выделяет наиболее сложные звуки в скороговорке
func highlightDifficultSounds(text string) {
	text = strings.ToLower(text)
	
	// Группы сложных звуков
	difficultGroups := map[string][]rune{
		"Шипящие":   {'ж', 'ш', 'щ', 'ч'},
		"Свистящие": {'с', 'з', 'ц'},
		"Сонорные":  {'р', 'л'},
		"Взрывные":  {'п', 'б', 'т', 'д', 'к', 'г'},
	}
	
	foundGroups := make(map[string]int)
	
	// Подсчитываем количество звуков каждой группы
	for groupName, sounds := range difficultGroups {
		count := 0
		for _, char := range text {
			for _, sound := range sounds {
				if char == sound {
					count++
					break
				}
			}
		}
		if count > 0 {
			foundGroups[groupName] = count
		}
	}
	
	// Если нашли сложные звуки, выводим рекомендации
	if len(foundGroups) > 0 {
		fmt.Println("\nСложные звуковые группы в этой скороговорке:")
		for group, count := range foundGroups {
			fmt.Printf("- %s (%d звуков)\n", group, count)
		}
	}
}

// updateUserPerformance обновляет статистику пользователя
func updateUserPerformance(profile *UserPerformance, twister TongueTwister, score int, focusArea int) {
	// Обновляем успешность по типам звуков
	difficulty := getDifficultyLevel(twister.Score)
	
	// Инициализируем значение, если его ещё нет
	if _, exists := profile.DifficultyRating[difficulty]; !exists {
		profile.DifficultyRating[difficulty] = 3.0 // Начальное среднее значение
	}
	
	// Обновляем статистику по сложности
	profile.DifficultyRating[difficulty] = (profile.DifficultyRating[difficulty]*0.7 + float64(score)*0.3)
	
	// Специфичные обновления в зависимости от фокуса
	if focusArea == 0 { // Артикуляция
		text := strings.ToLower(twister.Text)
		
		// Инициализируем значения, если их ещё нет
		if _, exists := profile.SuccessRate["шипящие"]; !exists {
			profile.SuccessRate["шипящие"] = 3.0
		}
		if _, exists := profile.SuccessRate["свистящие"]; !exists {
			profile.SuccessRate["свистящие"] = 3.0
		}
		if _, exists := profile.SuccessRate["сонорные"]; !exists {
			profile.SuccessRate["сонорные"] = 3.0
		}
		
		// Проверяем наличие сложных звуков
		if containsAny(text, []rune{'ш', 'щ', 'ж', 'ч'}) {
			profile.SuccessRate["шипящие"] = (profile.SuccessRate["шипящие"]*0.7 + float64(score)*0.3)
		}
		if containsAny(text, []rune{'с', 'з', 'ц'}) {
			profile.SuccessRate["свистящие"] = (profile.SuccessRate["свистящие"]*0.7 + float64(score)*0.3)
		}
		if containsAny(text, []rune{'р', 'л'}) {
			profile.SuccessRate["сонорные"] = (profile.SuccessRate["сонорные"]*0.7 + float64(score)*0.3)
		}
	}
	
	// Обновляем средний балл
	totalScore := 0
	for _, s := range profile.LastScores {
		totalScore += s
	}
	profile.AverageScore = float64(totalScore) / float64(len(profile.LastScores))
}

// adjustDifficulties корректирует сложность последующих раундов в зависимости от успешности
func adjustDifficulties(difficulties []float64, currentRound, score int) []float64 {
	adjustmentFactor := 1.0
	
	// Определяем фактор корректировки в зависимости от успешности
	switch score {
	case 1:
		adjustmentFactor = 0.8 // Значительно снижаем сложность при низкой оценке
	case 2:
		adjustmentFactor = 0.9 // Немного снижаем сложность
	case 3:
		adjustmentFactor = 1.0 // Оставляем как есть
	case 4:
		adjustmentFactor = 1.1 // Немного повышаем сложность
	case 5:
		adjustmentFactor = 1.2 // Значительно повышаем сложность
	}
	
	// Корректируем сложность последующих раундов
	for i := currentRound; i < len(difficulties); i++ {
		difficulties[i] *= adjustmentFactor
		// Ограничиваем сложность разумными пределами
		if difficulties[i] < 1.0 {
			difficulties[i] = 1.0
		} else if difficulties[i] > 5.0 {
			difficulties[i] = 5.0
		}
	}
	
	return difficulties
}

// provideFeedback дает обратную связь на основе оценки пользователя
func provideFeedback(score int, twister TongueTwister, focusArea int) {
	fmt.Println()
	
	// Общая обратная связь по оценке
	switch score {
	case 1, 2:
		fmt.Println("Не расстраивайтесь, эта скороговорка действительно непростая!")
		fmt.Println("Попробуйте разбить ее на маленькие части и проговорить медленнее.")
	case 3:
		fmt.Println("Неплохо! Продолжайте работать над дикцией.")
		fmt.Println("Обратите внимание на правильное положение языка и губ.")
	case 4:
		fmt.Println("Хорошо! Вы почти достигли совершенства.")
		fmt.Println("Попробуйте слегка увеличить скорость произношения.")
	case 5:
		fmt.Println("Отлично! Идеальное произношение!")
		fmt.Printf("Скороговорка \"%s\" сложности полностью освоена.\n", getDifficultyLevel(twister.Score))
	}
	
	// Дополнительная обратная связь в зависимости от фокуса
	if score < 4 {
		switch focusArea {
		case 0: // Артикуляция
			if twister.Stats.DifficultSounds > 0 {
				fmt.Println("▶ Совет: Уделите особое внимание чёткому произношению сложных звуков.")
			}
		case 1: // Ритм
			fmt.Println("▶ Совет: Попробуйте прохлопать ритм скороговорки перед произнесением.")
		case 2: // Ударения
			fmt.Println("▶ Совет: Произнесите скороговорку медленно, выделяя ударные слоги.")
		case 3: // Дыхание
			fmt.Println("▶ Совет: Сделайте несколько глубоких вдохов перед произнесением.")
		case 4: // Скорость
			fmt.Println("▶ Совет: Начните очень медленно и постепенно ускоряйтесь.")
		}
	}
}

// analyzeTrainingResults анализирует результаты тренировки и дает рекомендации
func analyzeTrainingResults(profile *UserPerformance, totalScore, totalRounds, focusArea int) {
	avgScore := float64(totalScore) / float64(totalRounds)
	
	fmt.Println("=== Анализ результатов тренировки ===")
	fmt.Printf("Ваш средний балл: %.1f из 5.0\n", avgScore)
	
	// Общий анализ
	if avgScore < 3.0 {
		fmt.Println("\nРекомендация: Продолжайте практиковаться в этом режиме.")
		fmt.Println("Фокусируйтесь на более медленном и четком произношении.")
	} else if avgScore < 4.0 {
		fmt.Println("\nРекомендация: Вы готовы к небольшому повышению сложности.")
		fmt.Println("Попробуйте увеличить скорость или перейти к более сложным скороговоркам.")
	} else {
		fmt.Println("\nРекомендация: Отличный результат! Вы готовы к продвинутому уровню.")
		fmt.Println("Переходите к более сложным скороговоркам или повышайте темп произношения.")
	}
	
	// Анализ по конкретным областям
	fmt.Println("\nДетальный анализ вашей дикции:")
	
	// Вывод проблемных областей на основе статистики
	if len(profile.SuccessRate) > 0 {
		minRate := 5.0
		minKey := ""
		
		for key, rate := range profile.SuccessRate {
			if rate < minRate {
				minRate = rate
				minKey = key
			}
		}
		
		if minKey != "" && minRate < 3.5 {
			fmt.Printf("• Обратите особое внимание на произношение звуков группы «%s»\n", minKey)
		}
	}
	
	// Дополнительный совет в зависимости от фокуса
	switch focusArea {
	case 0: // Артикуляция
		fmt.Println("• Для улучшения артикуляции рекомендуется делать упражнения для губ и языка")
		fmt.Println("  перед тренировкой скороговорок")
	case 1: // Ритм
		fmt.Println("• Для улучшения ритма попробуйте тренироваться с метрономом")
	case 2: // Ударения
		fmt.Println("• Для работы над ударениями читайте вслух стихи с выраженным ритмом")
	case 3: // Дыхание
		fmt.Println("• Для развития дыхания рекомендуются регулярные дыхательные упражнения")
	case 4: // Скорость
		fmt.Println("• Для увеличения скорости речи тренируйтесь ежедневно, постепенно повышая темп")
	}
	
	// Рекомендация по переходу к другому фокусу
	suggestNextTrainingFocus(focusArea, avgScore)
}

// suggestArticulationFocus provides specific guidance for articulation practice
func suggestArticulationFocus(text string, round int) {
	text = strings.ToLower(text)
	
	switch round {
	case 1:
		fmt.Println("Фокус: чёткое произношение всех согласных")
	case 2:
		fmt.Println("Фокус: выделение шипящих и свистящих звуков (ш, щ, ж, с, з)")
	case 3:
		fmt.Println("Фокус: проработка сочетаний согласных")
	case 4:
		fmt.Println("Фокус: плавные переходы между всеми звуками")
	default:
		fmt.Println("Фокус: идеальное произношение всех звуков")
	}
	
	// Находим сложные группы звуков для выделения
	for _, combo := range difficultCombinations {
		if strings.Contains(text, combo) {
			fmt.Printf("Обратите особое внимание на сочетание \"%s\"\n", combo)
			break
		}
	}
}

// suggestRhythmFocus provides guidance for rhythm training
func suggestRhythmFocus(text string, round int) {
	words := strings.Fields(text)
	
	switch round {
	case 1:
		fmt.Println("Фокус: равномерное произношение каждого слога")
	case 2:
		fmt.Println("Фокус: правильные паузы между словами")
	case 3:
		fmt.Println("Фокус: плавный ритмический рисунок")
	default:
		fmt.Println("Фокус: естественный ритм с сохранением четкости")
	}
	
	// Показываем ритмическую структуру с выделением ударений
	fmt.Print("Схема: ")
	for i, word := range words {
		if i > 0 {
			fmt.Print(" | ")
		}
		syllables := countRussianSyllables(word)
		fmt.Print(strings.Repeat("•", syllables))
	}
	fmt.Println()
}

// highlightStressPatterns shows stress patterns in words
func highlightStressPatterns(text string) {
	// В реальной системе здесь был бы доступ к словарю ударений
	// Здесь мы просто рекомендуем обратить внимание
	fmt.Println("Обратите внимание на правильные ударения в многосложных словах")
	
	words := strings.Fields(text)
	longWords := []string{}
	
	for _, word := range words {
		if countRussianSyllables(word) > 2 {
			longWords = append(longWords, word)
		}
	}
	
	if len(longWords) > 0 {
		fmt.Print("Многосложные слова: ")
		for i, word := range longWords {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(word)
		}
		fmt.Println()
	}
}

// suggestBreathingPattern provides guidance on breathing during speech
func suggestBreathingPattern(text string, round int) {
	words := strings.Fields(text)
	
	switch round {
	case 1:
		fmt.Println("Фокус: глубокий вдох перед началом")
	case 2:
		fmt.Println("Фокус: произнесите на одном дыхании")
	case 3:
		fmt.Println("Фокус: контроль интенсивности выдоха")
	default:
		fmt.Println("Фокус: плавное распределение дыхания")
	}
	
	// Рекомендуем места для вдоха при длинных фразах
	if len(words) > 5 {
		breathPoint := len(words) / 2
		
		fmt.Print("Рекомендация для дыхания: ")
		for i, word := range words {
			if i == breathPoint {
				fmt.Print("(вдох) ")
			}
			fmt.Print(word + " ")
		}
		fmt.Println()
	}
}

// suggestSpeedFocus provides guidance for speed training
func suggestSpeedFocus(wordCount int, round, totalRounds int) {
	// Расчет рекомендуемого темпа в словах в минуту
	baseSpeed := 60 // Базовая скорость 60 слов в минуту
	maxSpeed := 120 // Максимальная скорость 120 слов в минуту
	
	speedMultiplier := float64(round) / float64(totalRounds)
	targetSpeed := baseSpeed + int(float64(maxSpeed-baseSpeed)*speedMultiplier)
	
	fmt.Printf("Рекомендуемая скорость: примерно %d слов в минуту\n", targetSpeed)
	
	// Рекомендации по технике для текущего раунда
	switch round {
	case 1:
		fmt.Println("Фокус: четкое произношение в медленном темпе")
	case 2:
		fmt.Println("Фокус: постепенное увеличение темпа")
	case 3:
		fmt.Println("Фокус: плавность и скорость")
	default:
		fmt.Println("Фокус: максимальная скорость с сохранением четкости")
	}
	
	// Рассчитываем примерный интервал времени
	seconds := float64(wordCount) / (float64(targetSpeed) / 60.0)
	fmt.Printf("Целевое время: около %.1f секунд\n", seconds)
}

// suggestNextTrainingFocus recommends the next training focus based on current results
func suggestNextTrainingFocus(currentFocus int, score float64) {
	fmt.Println("\nРекомендации для будущих тренировок:")
	
	// Определяем статус пользователя на основе оценки
	var status string
	switch {
	case score < 2.5:
		status = "начинающий"
		fmt.Println("- Ваш текущий уровень: Начинающий")
		fmt.Println("- Рекомендуется больше практики в текущем режиме")
		fmt.Println("- Попробуйте снизить уровень сложности скороговорок")
		return
	case score < 3.5:
		status = "средний"
		fmt.Println("- Ваш текущий уровень: Средний")
		fmt.Println("- Вы делаете успехи, но требуется дополнительная практика")
		if rand.Float64() > 0.7 { // 30% шанс предложить следующий режим
			fmt.Println("- Можно попробовать перейти к следующему режиму тренировки")
		} else {
			fmt.Println("- Рекомендуется закрепить навыки в текущем режиме")
			return
		}
	case score < 4.5:
		status = "продвинутый"
		fmt.Println("- Ваш текущий уровень: Продвинутый")
		fmt.Println("- Отличные результаты! Готовы к новым вызовам")
	default:
		status = "эксперт"
		fmt.Println("- Ваш текущий уровень: Эксперт")
		fmt.Println("- Превосходно! Вы достигли высокого уровня мастерства")
	}
	
	// Оцениваем прогресс и даём более персонализированные рекомендации
	fmt.Printf("- Ваш средний балл: %.1f/5.0\n", score)
	
	// Предлагаем переход к следующей логичной области фокуса с учетом текущего статуса
	switch currentFocus {
	case 0: // После артикуляции
		fmt.Println("\nВозможные следующие шаги:")
		fmt.Println("- Попробуйте режим тренировки ритма (-focus=1)")
		fmt.Println("  Это поможет закрепить чёткое произношение в потоке речи")
		if status == "продвинутый" || status == "эксперт" {
			fmt.Println("- Поэкспериментируйте с более высоким уровнем сложности (-level=" + fmt.Sprintf("%d", min(5, int(score)+1)) + ")")
			fmt.Println("  или увеличьте число скороговорок в сессии (-count=10)")
		}
		
	case 1: // После ритма
		fmt.Println("\nВозможные следующие шаги:")
		fmt.Println("- Попробуйте режим тренировки ударений (-focus=2)")
		fmt.Println("  Это добавит выразительности вашей речи")
		if status == "продвинутый" || status == "эксперт" {
			fmt.Println("- Смешайте разные сложности скороговорок (-mix=true)")
			fmt.Println("  для более разнообразной тренировки")
		}
		
	case 2: // После ударений
		fmt.Println("\nВозможные следующие шаги:")
		fmt.Println("- Попробуйте режим тренировки дыхания (-focus=3)")
		fmt.Println("  Это поможет контролировать более длинные фразы")
		if status == "продвинутый" || status == "эксперт" {
			fmt.Println("- Попробуйте режим timed (-mode=timed) для тренировки")
			fmt.Println("  с ограничением по времени")
		}
		
	case 3: // После дыхания
		fmt.Println("\nВозможные следующие шаги:")
		fmt.Println("- Попробуйте режим тренировки скорости (-focus=4)")
		fmt.Println("  Теперь вы готовы увеличивать темп речи без потери качества")
		if status == "эксперт" {
			fmt.Println("- Переходите к режиму challenge (-mode=challenge)")
			fmt.Println("  для испытания ваших навыков в стрессовой ситуации")
		}
		
	case 4: // После скорости
		fmt.Println("\nВозможные следующие шаги:")
		if status == "эксперт" {
			fmt.Println("- Поздравляем! Вы достигли высочайшего уровня мастерства")
			fmt.Println("- Попробуйте комбинировать различные режимы тренировки:")
			fmt.Println("  * Сложные скороговорки (-difficulty=hard или -difficulty=expert)")
			fmt.Println("  * Максимальный уровень перфекции (-level=5)")
			fmt.Println("  * Увеличенное количество повторений (-reps=5)")
		} else {
			fmt.Println("- Попробуйте вернуться к тренировке артикуляции (-focus=0) на более высоком уровне")
			fmt.Println("  Это позволит достичь профессионального уровня дикции")
			fmt.Println("- Также рекомендуется попробовать тренировать все аспекты речи")
			fmt.Println("  поочередно для гармоничного развития")
		}
	}
	
	// Случайный дополнительный совет для разнообразия
	randomTips := []string{
		"Записывайте свою речь на диктофон для анализа произношения",
		"Читайте вслух стихи и прозу для общего развития дикции",
		"Выполняйте упражнения для губ и языка перед тренировкой",
		"Практикуйтесь каждый день по 15-20 минут для устойчивого прогресса",
		"Попробуйте разные темпы и интонации при произнесении скороговорок",
	}
	
	if rand.Float64() > 0.3 { // 70% шанс дать дополнительный совет
		randTip := randomTips[rand.Intn(len(randomTips))]
		fmt.Println("\n💡 Дополнительный совет: " + randTip)
	}
	
	// Финальный мотивирующий комментарий
	motivationalEndings := []string{
		"Успехов в совершенствовании дикции!",
		"Продолжайте практиковаться, и результаты не заставят себя ждать!",
		"Помните: регулярность важнее интенсивности!",
		"Даже профессиональные дикторы тренируются каждый день!",
	}
	
	fmt.Println("\n" + motivationalEndings[rand.Intn(len(motivationalEndings))])
}

// printComplexSounds highlights and prints the most challenging sounds in a tongue twister
func printComplexSounds(text string) {
	text = strings.ToLower(text)
	
	// Ищем сначала самые сложные сочетания
	foundCombos := []string{}
	
	for _, combo := range difficultCombinations {
		if strings.Contains(text, combo) {
			foundCombos = append(foundCombos, combo)
			if len(foundCombos) >= 3 {
				break
			}
		}
	}
	
	// Затем ищем отдельные сложные звуки
	foundSounds := make(map[rune]bool)
	
	for _, char := range text {
		for _, sound := range difficultSounds {
			if char == sound {
				foundSounds[char] = true
				break
			}
		}
	}
	
	// Выводим информацию
	if len(foundCombos) > 0 {
		fmt.Print(strings.Join(foundCombos, ", "))
		
		if len(foundSounds) > 0 {
			fmt.Print("; также звуки: ")
		}
	}
	
	if len(foundSounds) > 0 {
		soundList := ""
		for sound := range foundSounds {
			if len(soundList) > 0 {
				soundList += ", "
			}
			soundList += string(sound)
		}
		fmt.Print(soundList)
	}
	
	if len(foundCombos) == 0 && len(foundSounds) == 0 {
		fmt.Print("обычные звуки")
	}
	
	fmt.Println()
}

// printRhythmicStructure shows the rhythmic pattern of a tongue twister
func printRhythmicStructure(text string) {
	words := strings.Fields(text)
	rhythm := ""
	
	for i, word := range words {
		if i > 0 {
			rhythm += " "
		}
		
		// Simplified rhythm analysis - just show syllable count
		syllables := countRussianSyllables(word)
		rhythm += strings.Repeat("•", syllables)
	}
	
	fmt.Println(rhythm)
}

// countRussianSyllables estimates the number of syllables in a Russian word
func countRussianSyllables(word string) int {
	count := 0
	for _, char := range strings.ToLower(word) {
		if isRussianVowel(char) {
			count++
		}
	}
	
	// In case no vowels were found (unlikely in Russian)
	if count == 0 {
		return 1
	}
	
	return count
}

// selectBalancedTwisters selects twisters from different difficulty levels
func selectBalancedTwisters(easy, medium, hard, expert []TongueTwister, totalCount int) []TongueTwister {
	result := []TongueTwister{}
	
	// Calculate how many from each category to take
	// We want at least one from each non-empty category, then distribute the rest
	nonEmptyCategories := 0
	if len(easy) > 0 {
		nonEmptyCategories++
	}
	if len(medium) > 0 {
		nonEmptyCategories++
	}
	if len(hard) > 0 {
		nonEmptyCategories++
	}
	if len(expert) > 0 {
		nonEmptyCategories++
	}
	
	if nonEmptyCategories == 0 {
		return result
	}
	
	// Default distribution ratios for different categories
	easyRatio := 0.25
	mediumRatio := 0.30
	hardRatio := 0.30
	expertRatio := 0.15
	
	// Calculate initial counts
	easyCount := int(math.Round(float64(totalCount) * easyRatio))
	mediumCount := int(math.Round(float64(totalCount) * mediumRatio))
	hardCount := int(math.Round(float64(totalCount) * hardRatio))
	expertCount := int(math.Round(float64(totalCount) * expertRatio))
	
	// Ensure we don't exceed totalCount
	for easyCount + mediumCount + hardCount + expertCount > totalCount {
		if expertCount > 0 {
			expertCount--
		} else if hardCount > 0 {
			hardCount--
		} else if mediumCount > 0 {
			mediumCount--
		} else if easyCount > 0 {
			easyCount--
		}
	}
	
	// Add any missing counts
	missing := totalCount - (easyCount + mediumCount + hardCount + expertCount)
	for i := 0; i < missing; i++ {
		if len(medium) > mediumCount {
			mediumCount++
		} else if len(easy) > easyCount {
			easyCount++
		} else if len(hard) > hardCount {
			hardCount++
		} else if len(expert) > expertCount {
			expertCount++
		}
	}
	
	// Adjust if we have fewer twisters than requested
	easyCount = min(easyCount, len(easy))
	mediumCount = min(mediumCount, len(medium))
	hardCount = min(hardCount, len(hard))
	expertCount = min(expertCount, len(expert))
	
	// Select from each category
	if easyCount > 0 {
		result = append(result, selectRandomTwisters(easy, easyCount)...)
		fmt.Printf("Выбрано %d легких скороговорок\n", easyCount)
	}
	
	if mediumCount > 0 {
		result = append(result, selectRandomTwisters(medium, mediumCount)...)
		fmt.Printf("Выбрано %d средних скороговорок\n", mediumCount)
	}
	
	if hardCount > 0 {
		result = append(result, selectRandomTwisters(hard, hardCount)...)
		fmt.Printf("Выбрано %d сложных скороговорок\n", hardCount)
	}
	
	if expertCount > 0 {
		result = append(result, selectRandomTwisters(expert, expertCount)...)
		fmt.Printf("Выбрано %d очень сложных скороговорок\n", expertCount)
	}
	
	// Shuffle the final selection to mix difficulties
	shuffled := make([]TongueTwister, len(result))
	copy(shuffled, result)
	
	// Fisher-Yates shuffle
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	
	return shuffled
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
} 