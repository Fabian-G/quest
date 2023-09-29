Color when created on a full moon:
```
{ if = 'int(shell("jq -r .creation\ \|\ sub\(\"-\"\;\"\"\;\"g\"\)+\"00\" | xargs pom | sed s/\[^0-9\]//g | tr -d \"\\n\"", ""), 0) >= 94', fg = "#91a3b0" }
```