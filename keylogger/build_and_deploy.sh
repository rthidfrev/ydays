#!/bin/bash
PROGRAM_NAME="keylogger"
WINDOWS_DESKTOP="/mnt/c/Users/germo/Desktop"
cargo build --release --target x86_64-pc-windows-gnu
cp target/x86_64-pc-windows-gnu/release/$PROGRAM_NAME.exe $WINDOWS_DESKTOP/
echo "✅ Compilation terminée ! Fichier copié sur le Bureau Windows."
