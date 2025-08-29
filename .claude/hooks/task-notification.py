#!/usr/bin/env python3
import json
import sys
import os
import subprocess
import time
import platform

def get_notification_config():
    """Load notification configuration from config file"""
    config_file = '.claude/config/notification.json'
    if not os.path.exists(config_file):
        # Try legacy audio config
        legacy_config_file = '.claude/config/audio-notification.json'
        if os.path.exists(legacy_config_file):
            return migrate_legacy_config(legacy_config_file)
        return None
        
    try:
        with open(config_file, 'r', encoding='utf-8') as f:
            return json.load(f)
    except Exception as e:
        return None

def migrate_legacy_config(legacy_path):
    """Migrate legacy audio config to new notification config"""
    try:
        with open(legacy_path, 'r', encoding='utf-8') as f:
            legacy_config = json.load(f)
        
        # Convert to new format
        new_config = {
            "notification_types": ["audio"] if legacy_config.get("enabled", True) else [],
            "cooldown_seconds": legacy_config.get("cooldown_seconds", 2),
            "desktop": {
                "enabled": False,
                "show_details": True
            },
            "audio": {
                "enabled": legacy_config.get("enabled", True),
                "success_sound": legacy_config.get("success_sound", "success.wav"),
                "error_sound": legacy_config.get("error_sound", "error.wav"),
                "default_sound": legacy_config.get("default_sound", "complete.wav"),
                "volume": legacy_config.get("volume", 70)
            }
        }
        return new_config
    except Exception as e:
        return None

def should_send_notification(config):
    """Check if we should send a notification based on cooldown"""
    if not config.get('notification_types'):
        return False
        
    cooldown = config.get('cooldown_seconds', 2)
    if cooldown <= 0:
        return True
        
    # Check last notification time
    last_file = '.claude/last-notification-time'
    if os.path.exists(last_file):
        try:
            with open(last_file, 'r') as f:
                last_time = float(f.read().strip())
            if time.time() - last_time < cooldown:
                return False
        except:
            pass
            
    # Update last notification time
    try:
        with open(last_file, 'w') as f:
            f.write(str(time.time()))
    except:
        pass
        
    return True

def analyze_tool_result(tool_result):
    """Analyze tool result to determine success/failure and extract info"""
    success = True
    tool_name = "任务"
    details = ""
    
    try:
        result_str = str(tool_result).lower()
        
        # Check for error indicators
        error_keywords = ['error', 'failed', 'exception', 'timeout', 'denied', 'not found']
        for keyword in error_keywords:
            if keyword in result_str:
                success = False
                break
        
        # Try to extract tool information
        if isinstance(tool_result, dict):
            if 'tool_name' in tool_result:
                tool_name = tool_result['tool_name']
            elif 'command' in tool_result:
                tool_name = f"命令执行"
            elif 'file' in result_str:
                tool_name = f"文件操作"
        
    except Exception as e:
        pass
    
    return success, tool_name, details

def send_desktop_notification(title, message, message_type="info"):
    """Send desktop notification across platforms"""
    try:
        system = platform.system().lower()
        
        if system == 'darwin':  # macOS
            script = f'display notification "{message}" with title "{title}"'
            subprocess.run(['osascript', '-e', script], 
                         check=False, 
                         stdout=subprocess.DEVNULL, 
                         stderr=subprocess.DEVNULL)
        elif system == 'linux':
            # Try notify-send first
            try:
                subprocess.run(['notify-send', title, message], 
                             check=True,
                             stdout=subprocess.DEVNULL, 
                             stderr=subprocess.DEVNULL)
            except (subprocess.CalledProcessError, FileNotFoundError):
                # Try zenity as fallback
                try:
                    notification_text = f"{title}\\n{message}"
                    subprocess.run(['zenity', '--notification', f'--text={notification_text}'], 
                                 check=True,
                                 stdout=subprocess.DEVNULL, 
                                 stderr=subprocess.DEVNULL)
                except (subprocess.CalledProcessError, FileNotFoundError):
                    return False
        else:  # Windows
            # Use PowerShell balloon notification
            ps_script = f"""
                Add-Type -AssemblyName System.Windows.Forms
                $balloon = New-Object System.Windows.Forms.NotifyIcon
                $balloon.Icon = [System.Drawing.SystemIcons]::Information
                $balloon.BalloonTipTitle = "{title}"
                $balloon.BalloonTipText = "{message}"
                $balloon.Visible = $true
                $balloon.ShowBalloonTip(3000)
                Start-Sleep -Seconds 1
                $balloon.Dispose()
            """
            subprocess.run(['powershell', '-Command', ps_script], 
                         check=False,
                         stdout=subprocess.DEVNULL, 
                         stderr=subprocess.DEVNULL)
        return True
    except Exception as e:
        return False

def get_sound_file(config, success):
    """Get appropriate sound file based on result"""
    audio_config = config.get('audio', {})
    
    if success:
        return audio_config.get('success_sound', audio_config.get('default_sound', 'complete.wav'))
    else:
        return audio_config.get('error_sound', audio_config.get('default_sound', 'complete.wav'))

def get_sound_path(sound_file):
    """Get the full path to the sound file"""
    if os.path.isabs(sound_file):
        return sound_file if os.path.exists(sound_file) else None
        
    # Check in project .claude/sounds directory
    project_sound = os.path.join('.claude', 'sounds', sound_file)
    if os.path.exists(project_sound):
        return project_sound
        
    # Check relative to hooks directory
    script_dir = os.path.dirname(os.path.abspath(__file__))
    embedded_sound = os.path.join(script_dir, '..', 'sounds', sound_file)
    if os.path.exists(embedded_sound):
        return embedded_sound
        
    return None

def send_audio_notification(config, success):
    """Send audio notification"""
    audio_config = config.get('audio', {})
    if not audio_config.get('enabled', False):
        return False
    
    sound_file = get_sound_file(config, success)
    sound_path = get_sound_path(sound_file)
    
    if not sound_path:
        return False
    
    try:
        system = platform.system().lower()
        
        if system == 'darwin':  # macOS
            subprocess.run(['afplay', sound_path], 
                         check=False,
                         stdout=subprocess.DEVNULL, 
                         stderr=subprocess.DEVNULL)
        elif system == 'linux':
            # Try different audio players
            players = ['aplay', 'paplay', 'play']
            for player in players:
                try:
                    subprocess.run([player, sound_path], 
                                 check=True,
                                 stdout=subprocess.DEVNULL, 
                                 stderr=subprocess.DEVNULL)
                    break
                except (subprocess.CalledProcessError, FileNotFoundError):
                    continue
        else:  # Windows
            ps_script = f'$sound = New-Object Media.SoundPlayer "{sound_path}"; $sound.PlaySync()'
            subprocess.run(['powershell', '-Command', ps_script], 
                         check=False,
                         stdout=subprocess.DEVNULL, 
                         stderr=subprocess.DEVNULL)
        return True
    except Exception as e:
        return False

def main():
    try:
        # Load configuration
        config = get_notification_config()
        if not config:
            sys.exit(0)  # No config, exit silently
            
        # Check if notifications should be sent
        if not should_send_notification(config):
            sys.exit(0)
            
        # Read tool use data from stdin
        try:
            input_data = json.load(sys.stdin)
        except:
            sys.exit(0)
            
        # Analyze the tool result
        success, tool_name, details = analyze_tool_result(input_data)
        
        # Prepare notification content
        if success:
            title = "Claude Helper - 任务完成"
            message = f"✅ {tool_name} 操作完成"
            message_type = "success"
        else:
            title = "Claude Helper - 任务失败"  
            message = f"❌ {tool_name} 操作失败"
            message_type = "error"
        
        # Send notifications based on configured types
        notification_types = config.get('notification_types', [])
        
        # Send desktop notification
        if 'desktop' in notification_types:
            desktop_config = config.get('desktop', {})
            if desktop_config.get('enabled', False):
                send_desktop_notification(title, message, message_type)
        
        # Send audio notification
        if 'audio' in notification_types:
            send_audio_notification(config, success)
            
    except Exception as e:
        # On any error, fail silently
        try:
            with open('.claude/notification-error.log', 'a', encoding='utf-8', errors='replace') as f:
                import traceback
                f.write(f"Task notification error: {type(e).__name__}: {str(e)}\\n")
                f.write(f"Traceback: {traceback.format_exc()}\\n")
        except:
            pass  # Ignore logging errors
    
    sys.exit(0)

if __name__ == '__main__':
    main()
