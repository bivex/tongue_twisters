# Easy Trainer

This is a command-line application for practicing Russian tongue twisters. It helps users improve their diction, rhythm, stress, breathing, and speech speed through various training modes.

## Features

- **Load Tongue Twisters**: Loads tongue twisters from a JSON file.
- **Difficulty Analysis**: Automatically analyzes and scores the difficulty of each tongue twister.
- **Multiple Training Modes**:
    - `StandardMode`: Practice tongue twisters one by one at your own pace.
    - `TimedMode`: Practice with a time limit for each twister.
    - `RepeatMode`: Repeat each tongue twister a specified number of times.
    - `ChallengeMode`: Practice with increasing speed.
    - `PerfectionMode`: (NEW) Focuses on specific aspects of diction (articulation, rhythm, stress, breathing, speed) with adaptive difficulty and personalized feedback.

## Usage

Run the application from your terminal.

```bash
go run main.go [flags]
```

### Flags

- `--json <path>`: Path to JSON file with tongue twisters (default: `tongue_twisters/all_twisters.json`).
- `--count <number>`: How many random tongue twisters to select for training (default: `5`).
- `--difficulty <level>`: Difficulty level to select twisters from (e.g., `easy`, `medium`, `hard`, `expert`, `all`). Default is `all`.
- `--mode <mode_name>`: Training mode to use. Available modes: `standard`, `timed`, `repeat`, `challenge`, `perfection` (default: `standard`).
- `--time <seconds>`: Seconds per tongue twister in `timed` mode (default: `30`).
- `--reps <number>`: Number of repetitions in `repeat` mode (default: `3`).
- `--focus <area_id>`: (Perfection Mode) Focus area for diction training (0-4).
    - `0`: Артикуляция (Articulation) - Clear pronunciation of each sound.
    - `1`: Ритм (Rhythm) - Even speech tempo.
    - `2`: Ударения (Stress) - Correct word stress.
    - `3`: Дыхание (Breathing) - Breath control during pronunciation.
    - `4`: Скорость (Speed) - Increasing speed without losing quality.
- `--level <perfection_level>`: (Perfection Mode) Perfection level (1-5, higher is more demanding) (default: `3`).
- `--mix <boolean>`: Mix different difficulty levels when selecting twisters (default: `true`).

### Examples

- **Standard training with 10 random twisters:**
  ```bash
  go run main.go --count 10
  ```

- **Timed training (20 seconds per twister) with hard twisters:**
  ```bash
  go run main.go --mode timed --time 20 --difficulty hard
  ```

- **Perfection training focusing on articulation at level 4:**
  ```bash
  go run main.go --mode perfection --focus 0 --level 4
  ```

- **Repeat training, 5 repetitions of easy twisters:**
  ```bash
  go run main.go --mode repeat --reps 5 --difficulty easy
  ```

## Development

### Project Structure

- `main.go`: Main application logic, including training modes and analysis functions.
- `tongue_twisters/all_twisters.json`: JSON file containing the tongue twisters data.

### Adding New Tongue Twisters

You can add more tongue twisters by editing the `tongue_twisters/all_twisters.json` file. Each entry should be a JSON object with `number`, `date`, and `text` fields.

```json
[
  {
    "number": "1",
    "date": "2023-01-01",
    "text": "Ехал Грека через реку..."
  },
  {
    "number": "2",
    "date": "2023-01-02",
    "text": "Карл у Клары украл кораллы..."
  }
]
```
