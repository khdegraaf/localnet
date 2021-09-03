# TMux Help

## Common keys

    Ctrl-b 0..9       Select window 0..9
    Ctrl-b p,n        Select previous, next window
    Ctrl-b w          Select window interactively
    Ctrl-b [          Enter scrollback/copy mode (ESC to exit)
      Ctrl-s to search in scrollback in this mode
    Ctrl-b d          Detach from the terminal (keep running in background)

## Dealing with windows 

    * Exited programs will be marked in status line.

    * Ctrl-C sends SIGINT. It will stop the program inside

    * Ctrl-\ sends SIGQUIT. Use it to dump goroutines with stacktraces.

    * Press Enter to restart exited program again.