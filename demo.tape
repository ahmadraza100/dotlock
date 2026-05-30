Output demo.gif

Set FontSize 14
Set Width 1400
Set Height 800
Set Theme "Dracula"
Set Framerate 24
Set PlaybackSpeed 0.8

Type "cd /tmp && rm -rf demo && mkdir demo && cd demo && dotlock init"
Enter
Sleep 800ms

Type "dotlock set DATABASE_URL"
Enter
Sleep 200ms
Type "postgres://localhost:5432/myapp"
Enter
Sleep 400ms

Type "dotlock set STRIPE_KEY"
Enter
Sleep 200ms
Type "sk_live_9xKqW2mNpLvTjRbYc"
Enter
Sleep 400ms

Type "dotlock set JWT_SECRET"
Enter
Sleep 200ms
Type "supersecretjwtsigningkey"
Enter
Sleep 400ms

Type "dotlock ui"
Enter
Sleep 3s

Tab
Sleep 800ms

Down
Sleep 500ms

Type "v"
Sleep 4s

Type "q"
Sleep 500ms