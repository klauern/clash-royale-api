"""Development watcher script for auto-formatting on save."""
import sys
import time

try:
    from watchdog.observers import Observer
    from watchdog.events import FileSystemEventHandler

    class ChangeHandler(FileSystemEventHandler):
        def on_modified(self, event):
            if event.src_path.endswith('.py'):
                print(f'Changed: {event.src_path}')
                # Auto-format on save
                import subprocess
                subprocess.run(['uv', 'run', 'python', '-m', 'black', event.src_path], capture_output=True)

    observer = Observer()
    observer.schedule(ChangeHandler(), path='src/', recursive=True)
    observer.start()

    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        observer.stop()
    observer.join()
except ImportError:
    print("Watchdog not installed. Run: uv add --dev watchdog")
    sys.exit(1)