# heimdallr-sense

> *"Hearing the voice before the storm"*

Voice Activity Detection на Go с микрофона через PipeWire (`pw-cat`). Чистый Go, без CGO.

Захватывает аудио с микрофона, прогоняет через WebRTC VAD, печатает `[VOICE START]` / `[VOICE END]` и записывает аудио в WAV-файлы или отправляет на сервер.

## Зависимости

- Linux с PipeWire (`pw-cat`)
- Go 1.25+

## Сборка

```bash
make                    # собрать все архитектуры
make all                # то же самое
make clean              # удалить бинарники
```

Собранные бинарники в `build/`:

| Бинарник | Архитектура |
|---|---|
| `heimdallr-sense-linux-amd64` | x86_64 |
| `heimdallr-sense-linux-arm64` | ARM64 (Raspberry Pi 4+, Rock5) |
| `heimdallr-sense-linux-armv7` | ARMv7 (Tinker Board, RPi 2/3) |

Локальная сборка:

```bash
go build -o heimdallr-sense ./cmd/vad
```

## Запуск

```bash
./heimdallr-sense
```

Пример вывода (`log_enabled: true`):

```json
{"time":"2026-07-23T15:30:45Z","level":"INFO","msg":"config loaded from ./config.yaml"}
{"time":"2026-07-23T15:30:45Z","level":"INFO","msg":"listening","rate":8000,"frame_ms":30,"frames":5,"voice_threshold":4,"silence_threshold":3,"record_mode":"file"}
{"time":"2026-07-23T15:30:46Z","level":"INFO","msg":"voice start"}
{"time":"2026-07-23T15:30:46Z","level":"INFO","msg":"recording started","pre_buffer_chunks":3}
{"time":"2026-07-23T15:30:48Z","level":"INFO","msg":"voice end"}
{"time":"2026-07-23T15:30:48Z","level":"INFO","msg":"recording saved","chunks":8,"duration_s":"1.2"}
{"time":"2026-07-23T15:30:48Z","level":"INFO","msg":"file saved","path":"./recordings/2026-07-23_15-30-46.000.wav"}
```

Ctrl+C для остановки.

## Конфигурация

Все параметры задаются в `config.yaml`. Конфиг ищется в порядке:

1. `/etc/heimdallr-sense/config.yaml`
2. `./config.yaml` (текущая директория)

Если файл не найден — используются дефолты.

```yaml
sample_rate: 8000
vad_frame_ms: 30
frames_per_chunk: 5
voice_threshold: 4
silence_threshold: 3
vad_mode: 3

# audio_source: pw-cat, arecord, custom
audio_source: pw-cat
audio_command: ""

# recording: none, file, https, both
record_mode: file
record_dir: ./recordings
pre_buffer_chunks: 3
https_url: ""
http_timeout: 10
min_chunks: 3
tls_skip_verify: false
log_enabled: true
```

### Параметры VAD

| Параметр | Описание | Дефолт |
|---|---|---|
| `sample_rate` | Частота дискретизации микрофона (Hz) | 16000 |
| `vad_frame_ms` | Длительность фрейма VAD (мс) | 30 |
| `frames_per_chunk` | Сколько фреймов анализировать за раз | 5 |
| `voice_threshold` | Мин. фреймов с голосом для START | 4 |
| `silence_threshold` | Мин. чанков без голоса для END | 3 |
| `vad_mode` | Агрессивность VAD (0-3) | 3 |

### Параметры аудио

| Параметр | Описание | Дефолт |
|---|---|---|
| `audio_source` | Источник аудио: `pw-cat`, `arecord`, `custom` | pw-cat |
| `audio_command` | Кастомная команда (при `audio_source: custom`) | "" |

Примеры:

```yaml
# PipeWire (по умолчанию)
audio_source: pw-cat

# ALSA
audio_source: arecord

# Кастомная команда (например, через SSH)
audio_source: custom
audio_command: "ssh remote-host arecord -f S16_LE -r 8000 -c 1 -t raw -"
```

### Параметры записи

| Параметр | Описание | Дефолт |
|---|---|---|
| `record_mode` | Режим записи: `none`, `file`, `https`, `both` | none |
| `record_dir` | Каталог для WAV-файлов | ./recordings |
| `pre_buffer_chunks` | Чанков в кольцевом буфере (запись до голоса) | 3 |
| `https_url` | URL для POST-запроса (multipart/form-data) | "" |
| `http_timeout` | Таймаут HTTP-запроса (сек) | 10 |
| `min_chunks` | Мин. чанков для сохранения записи (отсекает короткие шумы) | 3 |
| `tls_skip_verify` | Игнорировать проверку TLS-сертификата | false |
| `log_enabled` | Включить JSON-логи (slog) | true |

### Расчёт окна

- Длительность чанка = `vad_frame_ms` * `frames_per_chunk` (по умолчанию 150мс)
- Задержка VOICE START = 1 чанк (150мс)
- Задержка VOICE END = `silence_threshold` чанков (450мс)
- Pre-buffer = `pre_buffer_chunks` чанков (450мс) — аудио до начала голоса

### Режимы VAD

| Значение | Режим | Описание |
|---|---|---|
| 0 | Quality | Максимальная чувствительность |
| 1 | LowBitrate | Средняя |
| 2 | Aggressive | Снижает ложные срабатывания |
| 3 | VeryAggressive | Минимальная чувствительность, только явная речь |

### Режимы записи

| Значение | Описание |
|---|---|
| `none` | Запись отключена |
| `file` | Сохранение WAV в `record_dir` |
| `https` | POST на `https_url` |
| `both` | И файл, и POST |

### Формат WAV-файлов

- 16-bit PCM, mono
- Имя файла: `YYYY-MM-DD_HH-MM-SS.mmm.wav` (время начала записи)
- Запись включает pre-buffer (аудио до детекции голоса)

### HTTPS-загрузка

Файл отправляется как `multipart/form-data` с полем `file`. Пример серверного обработчика:

```python
@app.post("/upload")
async def upload(file: UploadFile):
    data = await file.read()
    # data — WAV-файл
```

## Запуск как сервис

### systemd

Создай файл `/etc/systemd/system/heimdallr-sense.service`:

```ini
[Unit]
Description=Heimdallr Sense - Voice Activity Detection
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/heimdallr-sense
WorkingDirectory=/etc/heimdallr
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

```bash
sudo cp build/heimdallr-sense-linux-amd64 /usr/local/bin/heimdallr-sense
sudo mkdir -p /etc/heimdallr-sense
sudo cp config.yaml /etc/heimdallr-sense/config.yaml
sudo cp heimdallr-sense.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now heimdallr-sense

# логи
journalctl -u heimdallr-sense -f
```

### SystemV (init.d)

Создай файл `/etc/init.d/heimdallr-sense`:

```bash
#!/bin/sh
### BEGIN INIT INFO
# Provides:          heimdallr-sense
# Required-Start:    $local_fs $network
# Required-Stop:     $local_fs $network
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Voice Activity Detection
# Description:       Heimdallr Sense VAD service
### END INIT INFO

NAME=heimdallr-sense
DAEMON=/usr/local/bin/$NAME
CONFIG=/etc/heimdallr-sense/config.yaml
PIDFILE=/var/run/$NAME.pid
LOGFILE=/var/log/$NAME.log

case "$1" in
    start)
        echo "Starting $NAME"
        nohup $DAEMON -config $CONFIG >> $LOGFILE 2>&1 &
        echo $! > $PIDFILE
        ;;
    stop)
        if [ -f $PIDFILE ]; then
            echo "Stopping $NAME"
            kill $(cat $PIDFILE)
            rm $PIDFILE
        fi
        ;;
    restart)
        $0 stop
        sleep 1
        $0 start
        ;;
    *)
        echo "Usage: $0 {start|stop|restart}"
        exit 1
        ;;
esac
exit 0
```

```bash
sudo cp build/heimdallr-sense-linux-amd64 /usr/local/bin/heimdallr-sense
sudo mkdir -p /etc/heimdallr-sense
sudo cp config.yaml /etc/heimdallr-sense/config.yaml
sudo cp heimdallr-sense.init /etc/init.d/heimdallr-sense
sudo chmod +x /etc/init.d/heimdallr-sense
sudo update-rc.d heimdallr-sense defaults

# ручное управление
sudo service heimdallr-sense start
sudo service heimdallr-sense stop
```

## Архитектура

```
pw-cat (микрофон)
  → raw S16LE PCM
    → ring buffer (pre-buffer)
      → WebRTC VAD → [VOICE START/END]
        → recording buffer → WAV file / HTTPS POST
```

Программа запускает `pw-cat -r --format s16 --rate <rate> --channels 1 -`, читает сырые PCM-данные, разбивает на фреймы и прогоняет через WebRTC VAD (реализация [rolandhe/go-vad](https://github.com/rolandhe/go-vad) — чистый Go порт libfvad).

При обнаружении голоса буфер pre-buffer сбрасывается в начало записи, аудио копится пока голос есть, и после `silence_threshold` чанков тишины запись финализируется.
