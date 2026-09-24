from pathlib import Path
import os
import subprocess


SERVER_DIR = Path(__file__).resolve().parent
SERVER_BINARY = SERVER_DIR / "server"
REMOTE_BINARY = "codeka.com:/home/web/snipit/server"


def main() -> None:
    build_environment = os.environ.copy()
    build_environment.update({"GOOS": "linux", "GOARCH": "amd64"})

    subprocess.run(
        ["go", "build", "-o", str(SERVER_BINARY), "."],
        cwd=SERVER_DIR,
        env=build_environment,
        check=True,
    )
    subprocess.run(
        ["scp", str(SERVER_BINARY), REMOTE_BINARY],
        cwd=SERVER_DIR,
        check=True,
    )


if __name__ == "__main__":
    main()