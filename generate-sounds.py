#!/usr/bin/env python3
"""
Generate simple audio notification sounds for Claude Helper

This script generates basic WAV audio files for testing purposes.
For production, consider using higher quality audio files.

Requirements: pip install numpy scipy
"""

import numpy as np
import os
from scipy.io.wavfile import write

# Audio parameters
SAMPLE_RATE = 44100
DURATION_SHORT = 0.3  # seconds
DURATION_MEDIUM = 0.5  # seconds  
DURATION_LONG = 0.8   # seconds

def generate_tone(frequency, duration, sample_rate=SAMPLE_RATE, amplitude=0.3):
    """Generate a simple sine wave tone"""
    t = np.linspace(0, duration, int(sample_rate * duration), False)
    # Apply fade in/out to avoid clicks
    fade_samples = int(sample_rate * 0.05)  # 50ms fade
    tone = amplitude * np.sin(2 * np.pi * frequency * t)
    
    # Apply fade in
    if len(tone) > fade_samples:
        fade_in = np.linspace(0, 1, fade_samples)
        tone[:fade_samples] *= fade_in
        
    # Apply fade out  
    if len(tone) > fade_samples:
        fade_out = np.linspace(1, 0, fade_samples)
        tone[-fade_samples:] *= fade_out
        
    return tone

def generate_chord(frequencies, duration, sample_rate=SAMPLE_RATE, amplitude=0.2):
    """Generate a chord from multiple frequencies"""
    chord = np.zeros(int(sample_rate * duration))
    for freq in frequencies:
        tone = generate_tone(freq, duration, sample_rate, amplitude)
        chord += tone
    return chord / len(frequencies)  # Normalize

def generate_bell(base_freq, duration, sample_rate=SAMPLE_RATE):
    """Generate bell-like sound with harmonics"""
    t = np.linspace(0, duration, int(sample_rate * duration), False)
    
    # Bell harmonics (simplified)
    harmonics = [1.0, 2.4, 5.4, 8.9, 13.3]
    amplitudes = [1.0, 0.6, 0.4, 0.25, 0.15]
    
    bell = np.zeros_like(t)
    for harmonic, amp in zip(harmonics, amplitudes):
        freq = base_freq * harmonic
        # Exponential decay for bell effect
        decay = np.exp(-t * 3)
        wave = amp * np.sin(2 * np.pi * freq * t) * decay
        bell += wave
        
    # Normalize
    bell = bell / np.max(np.abs(bell)) * 0.3
    return bell

def save_audio(filename, audio_data, sample_rate=SAMPLE_RATE):
    """Save audio data as WAV file"""
    # Convert to 16-bit integers
    audio_16bit = (audio_data * 32767).astype(np.int16)
    write(filename, sample_rate, audio_16bit)
    print(f"Generated: {filename}")

def main():
    """Generate all notification sounds"""
    output_dir = "internal/assets/sounds"
    os.makedirs(output_dir, exist_ok=True)
    
    print("Generating audio notification sounds...")
    
    # Success sound - Happy chord (C-E-G major)
    success_chord = generate_chord([523, 659, 784], DURATION_MEDIUM)  # C5-E5-G5
    save_audio(os.path.join(output_dir, "success.wav"), success_chord)
    
    # Error sound - Low warning tone
    error_tone = generate_tone(220, DURATION_SHORT)  # A3
    save_audio(os.path.join(output_dir, "error.wav"), error_tone)
    
    # Complete sound - Simple notification
    complete_tone = generate_tone(800, DURATION_SHORT)  # G#5
    save_audio(os.path.join(output_dir, "complete.wav"), complete_tone)
    
    # Attention sound - Two-tone alert  
    attention_part1 = generate_tone(800, 0.15)
    attention_part2 = generate_tone(600, 0.15)
    attention_gap = np.zeros(int(SAMPLE_RATE * 0.1))  # 100ms gap
    attention_sound = np.concatenate([attention_part1, attention_gap, attention_part2])
    save_audio(os.path.join(output_dir, "attention.wav"), attention_sound)
    
    # Subtle sound - Very soft tone
    subtle_tone = generate_tone(600, 0.2, amplitude=0.15)
    save_audio(os.path.join(output_dir, "subtle.wav"), subtle_tone)
    
    # Chime sound - High pitched clear tone
    chime_tone = generate_tone(1200, DURATION_MEDIUM, amplitude=0.25)
    save_audio(os.path.join(output_dir, "chime.wav"), chime_tone)
    
    # Bell sound - Bell with harmonics
    bell_sound = generate_bell(400, DURATION_LONG)
    save_audio(os.path.join(output_dir, "bell.wav"), bell_sound)
    
    print("\nAll audio files generated successfully!")
    print(f"Files saved in: {output_dir}")
    print("\nTo test the sounds:")
    print("macOS: afplay internal/assets/sounds/success.wav")
    print("Linux: aplay internal/assets/sounds/success.wav") 
    print("Windows: powershell -c \"(New-Object Media.SoundPlayer 'internal/assets/sounds/success.wav').PlaySync()\"")

if __name__ == "__main__":
    try:
        main()
    except ImportError as e:
        print("Error: Required packages not installed")
        print("Please install: pip install numpy scipy")
        print(f"Details: {e}")
    except Exception as e:
        print(f"Error generating sounds: {e}")