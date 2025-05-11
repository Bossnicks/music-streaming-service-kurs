import sys
import numpy as np
import librosa
import json
import warnings
from scipy.stats import norm


warnings.filterwarnings("ignore")

def extract_features(filepath):
    result = {}

    # Load with librosa
    y, sr = librosa.load(filepath, sr=44100)
    #prior = norm(loc=110, scale=40)  # Темп в пределах 60-70 BPM


    onset_env = librosa.onset.onset_strength(y=y, sr=sr, aggregate=np.median)

    # Duration
    duration = librosa.get_duration(y=y, sr=sr)
    result["duration_sec"] = round(duration, 2)

    # # Tempo
    # tempo = librosa.beat.tempo(y=y, sr=sr,onset_envelope=onset_env, hop_length=64, aggregate=None)
    # raw_tempo = np.mean(tempo)
    # result["tempo_bpm"] = round(float(raw_tempo), 2)

    # Tempo (raw estimate, fallback)
    tempo_raw = librosa.beat.tempo(y=y, sr=sr, onset_envelope=onset_env, hop_length=64, aggregate=None)
    raw_tempo = np.mean(tempo_raw)

    # Refined tempo using beat intervals
    _, beat_times = librosa.beat.beat_track(y=y, sr=sr, units='time')
    if len(beat_times) > 1:
        beat_intervals = np.diff(beat_times)
        avg_interval = np.mean(beat_intervals)
        refined_bpm = 60.0 / avg_interval
        result["tempo_bpm"] = round(refined_bpm, 2)
    else:
        result["tempo_bpm"] = round(float(raw_tempo), 2)

    # Chromagram (harmonic profile)
    chroma = librosa.feature.chroma_stft(y=y, sr=sr)
    result["chroma_mean"] = round(np.mean(chroma), 3)

    # # 4. Основной тональный ключ (0–11, где 0 = C, 1 = C#, ..., 11 = B)
    # key = int(np.argmax(np.mean(chroma, axis=1)))
    # result["key"] = key

    # # 5. Лад — мажор (major) или минор (minor)
    # tonnetz = librosa.feature.tonnetz(y=y, sr=sr)
    # result["mode"] = "major" if np.mean(tonnetz) > 0 else "minor"

    # Root Mean Square Energy (loudness)
    rmse = librosa.feature.rms(y=y)[0]
    result["rmse_mean"] = round(np.mean(rmse), 3)

    # Spectral centroid = brightness
    spectral_centroid = librosa.feature.spectral_centroid(y=y, sr=sr)[0]
    result["spectral_centroid"] = round(np.mean(spectral_centroid), 2)

    # Spectral bandwidth
    spectral_bandwidth = librosa.feature.spectral_bandwidth(y=y, sr=sr)[0]
    result["spectral_bandwidth"] = round(np.mean(spectral_bandwidth), 2)

    # Spectral rolloff
    rolloff = librosa.feature.spectral_rolloff(y=y, sr=sr)[0]
    result["rolloff"] = round(np.mean(rolloff), 2)

    # Zero Crossing Rate
    zcr = librosa.feature.zero_crossing_rate(y)[0]
    result["zero_crossing_rate"] = round(np.mean(zcr), 5)

    # MFCCs
    # mfcc = librosa.feature.mfcc(y=y, sr=sr, n_mfcc=128)
    # result["mfcc"] = [round(float(coeff), 3) for coeff in np.mean(mfcc, axis=1)]

    # Convert all result values to regular float type to ensure JSON serialization
    for key in result:
        result[key] = float(result[key]) if isinstance(result[key], np.float32) else result[key]

    return result

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Использование: python analyze_song.py audio_file.mp3")
        sys.exit(1)

    filepath = sys.argv[1]
    features = extract_features(filepath)
    print(json.dumps(features, indent=2))


# import librosa
# import numpy as np
# import json
# import sys
# import warnings
# from contextlib import redirect_stderr
# import io

# def analyze_audio(file_path):
#     """Анализ аудио с улучшенным определением темпа"""
#     stderr_buffer = io.StringIO()

#     try:
#         with warnings.catch_warnings():
#             warnings.simplefilter("ignore")
#             with redirect_stderr(stderr_buffer):
#                 y, sr = librosa.load(file_path, sr=None)

#                 # 1. Длительность
#                 duration = int(round(librosa.get_duration(y=y, sr=sr)))

#                 # 2. Улучшенное определение темпа
#                 tempo_librosa = int(round(librosa.beat.tempo(y=y, sr=sr)[0]))

#                 onset_frames = librosa.onset.onset_detect(y=y, sr=sr)
#                 onset_times = librosa.frames_to_time(onset_frames, sr=sr)
#                 if len(onset_times) > 1:
#                     ioi = np.diff(onset_times)
#                     median_ioi = np.median(ioi)
#                     tempo_ioi = int(round(60 / median_ioi))
#                 else:
#                     tempo_ioi = 0

#                 tempo = max(tempo_librosa, tempo_ioi)

#                 # 3. Гармония и тональность
#                 chroma = librosa.feature.chroma_stft(y=y, sr=sr)
#                 key = int(np.argmax(np.mean(chroma, axis=1)))
#                 mode = "major" if np.mean(librosa.feature.tonnetz(y=y, sr=sr)) > 0 else "minor"
#                 tonal_stability = int(100 - np.std(chroma) * 10)

#                 # 4. Тембр
#                 centroid = int(round(np.mean(librosa.feature.spectral_centroid(y=y, sr=sr))))
#                 roughness = int(round(np.mean(librosa.feature.spectral_flatness(y=y)) * 100))

#                 # 5. Ритм и бит
#                 onset_env = librosa.onset.onset_strength(y=y, sr=sr)
#                 danceability = int(round(np.mean(onset_env) * 50))
#                 beats = int(len(onset_frames))

#                 # 6. Динамика
#                 energy = int(round(np.mean(librosa.feature.rms(y=y)) * 100))
#                 vocals = int(round(np.mean(librosa.effects.harmonic(y)) * 100))

#                 return {
#                     "duration_sec": duration,
#                     "tempo_bpm": tempo,
#                     "key": key,
#                     "mode": mode,
#                     "brightness_hz": centroid,
#                     "danceability": min(100, max(0, danceability)),
#                     "energy": min(100, max(0, energy)),
#                     "roughness": min(100, max(0, roughness)),
#                     "tonal_stability": min(100, max(0, tonal_stability)),
#                     "beats": beats,
#                     "vocals": min(100, max(0, vocals))
#                 }

#     except Exception as e:
#         return {
#             "error": str(e),
#             "stderr": stderr_buffer.getvalue()
#         }

# if __name__ == "__main__":
#     if len(sys.argv) != 2:
#         print(json.dumps({"error": "Требуется 1 аргумент - путь к файлу"}))
#         sys.exit(1)

#     result = analyze_audio(sys.argv[1])
#     print(json.dumps(result, ensure_ascii=False, indent=2))

# import sys
# import numpy as np
# import librosa
# import json
# import warnings

# warnings.filterwarnings("ignore")

# def extract_features(filepath):
#     result = {}

#     y, sr = librosa.load(filepath, sr=44100)

#     # --- Основные аудиофичи ---
    
#     # 1. Длительность (в секундах)
#     duration = librosa.get_duration(y=y, sr=sr)
#     result["duration_sec"] = round(duration, 2)

#     # 2. Темп (в ударах в минуту — BPM)
#     onset_env = librosa.onset.onset_strength(y=y, sr=sr, aggregate=np.median)
#     tempo_raw = librosa.beat.tempo(y=y, sr=sr, onset_envelope=onset_env, hop_length=64, aggregate=None)
#     raw_tempo = np.mean(tempo_raw)
#     _, beat_times = librosa.beat.beat_track(y=y, sr=sr, units='time')
#     if len(beat_times) > 1:
#         beat_intervals = np.diff(beat_times)
#         avg_interval = np.mean(beat_intervals)
#         refined_bpm = 60.0 / avg_interval
#         result["tempo_bpm"] = round(refined_bpm, 2)
#     else:
#         result["tempo_bpm"] = round(float(raw_tempo), 2)

#     # 3. Хромограмма (характеристика гармонии и тональности)
#     chroma = librosa.feature.chroma_stft(y=y, sr=sr)
#     result["chroma_mean"] = round(np.mean(chroma), 3)

#     # 4. Основной тональный ключ (0–11, где 0 = C, 1 = C#, ..., 11 = B)
#     key = int(np.argmax(np.mean(chroma, axis=1)))
#     result["key"] = key

#     # 5. Лад — мажор (major) или минор (minor)
#     tonnetz = librosa.feature.tonnetz(y=y, sr=sr)
#     result["mode"] = "major" if np.mean(tonnetz) > 0 else "minor"

#     # 6. Стабильность тональности (0–100): насколько устойчивы ноты
#     result["tonal_stability"] = int(100 - np.std(chroma) * 10)

#     # 7. Энергия (громкость) — от 0 до 100
#     rmse = librosa.feature.rms(y=y)[0]
#     result["rmse_mean"] = round(np.mean(rmse), 3)
#     result["energy"] = min(100, max(0, int(round(np.mean(rmse) * 100))))

#     # 8. Центроид спектра (яркость звука) — в герцах
#     spectral_centroid = librosa.feature.spectral_centroid(y=y, sr=sr)[0]
#     result["spectral_centroid"] = round(np.mean(spectral_centroid), 2)
#     result["brightness_hz"] = int(round(np.mean(spectral_centroid)))

#     # 9. Спектральная ширина (насыщенность верхами)
#     spectral_bandwidth = librosa.feature.spectral_bandwidth(y=y, sr=sr)[0]
#     result["spectral_bandwidth"] = round(np.mean(spectral_bandwidth), 2)

#     # 10. Спектральный спад (rolloff) — до какой частоты сосредоточено 85% энергии сигнала
#     rolloff = librosa.feature.spectral_rolloff(y=y, sr=sr)[0]
#     result["rolloff"] = round(np.mean(rolloff), 2)

#     # 11. Грубость спектра (flatness) — 0 (гладкий) до 100 (шумный, шероховатый)
#     flatness = librosa.feature.spectral_flatness(y=y)[0]
#     result["roughness"] = min(100, max(0, int(round(np.mean(flatness) * 100))))

#     # 12. Частота пересечения нуля (Zero Crossing Rate) — мера шумности сигнала
#     zcr = librosa.feature.zero_crossing_rate(y)[0]
#     result["zero_crossing_rate"] = round(np.mean(zcr), 5)

#     # 13. Танцевальность (насколько выразительно ритм выражен) — 0–100
#     onset_env = librosa.onset.onset_strength(y=y, sr=sr)
#     result["danceability"] = min(100, max(0, int(round(np.mean(onset_env) * 50))))

#     # 14. Кол-во битов (ударов)
#     result["beats"] = int(len(librosa.onset.onset_detect(y=y, sr=sr)))

#     # 15. Доля гармонического сигнала (вокалы, мелодия)
#     vocals = librosa.effects.harmonic(y)
#     result["vocals"] = min(100, max(0, int(round(np.mean(vocals) * 100))))

#     # Приводим типы для сериализации в JSON
#     for key in result:
#         if isinstance(result[key], np.generic):
#             result[key] = float(result[key])

#     return result

# if __name__ == "__main__":
#     if len(sys.argv) != 2:
#         print("Использование: python analyze_song.py audio_file.mp3")
#         sys.exit(1)

#     filepath = sys.argv[1]
#     features = extract_features(filepath)
#     print(json.dumps(features, ensure_ascii=False, indent=2))