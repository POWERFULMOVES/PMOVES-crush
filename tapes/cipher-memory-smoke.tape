Output tapes/cipher-memory-smoke.gif
Set FontFamily "JetBrains Mono Nerd Font"
Set FontSize 16
Set Width 1200
Set Height 600
Set Padding 20
Set Theme "Dracula"
Set Shell "bash"

Type "# Cipher Memory smoke test"
Enter
Sleep 500ms

Type "make -C pmoves up-cipher-full"
Enter
Sleep 2s

Type "# Starting full Cipher embedding stack..."
Sleep 2s

Type "# Cipher Memory + embedding pipeline ready"
Sleep 1s

Type "make -C pmoves cipher-memory-smoke"
Enter
Sleep 2s

Type "# Testing cipher /api/memory POST..."
Sleep 1s

Type "# POST OK"
Sleep 500ms

Type "# Testing cipher /api/memory/search..."
Sleep 1s

Type "# SEARCH OK"
Sleep 500ms

Type "# ✔ Cipher memory pipeline smoke passed"
Sleep 1s
