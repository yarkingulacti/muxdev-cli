package doc

// staticPages returns curated wiki topics that are not tied to a single subcommand.
func staticPages() []Page {
	return []Page{
		{
			ID:       "welcome",
			Category: "Start here",
			Title:    "Welcome",
			Body: `muxdev runs your local dev stack from muxdev.yaml — pick services,
stream logs in a TUI, resolve port conflicts, and review past sessions.

Quick start
  1. cd your-project
  2. muxdev init          create muxdev.yaml (wizard)
  3. muxdev               pick services → run

No config yet? muxdev offers init automatically when started in a new repo.

Global flags (with any command)
  -c, --config PATH       muxdev.yaml location
      --no-interactive    plain stdout/stderr, no TUI
      --runtime MODE      sync (default) or async startup
      --focus IDS         comma-separated service ids

More
  muxdev list             services table for this project
  muxdev configure        edit manifest interactively
  muxdev logs             browse saved runtime sessions`,
			TryCommand: "muxdev version",
		},
		{
			ID:       "runtime-tui",
			Category: "Runtime",
			Title:    "Runtime TUI & keyboard",
			Body: `While services run, muxdev shows a live log panel.

Navigation
  PgUp / PgDown         scroll one line
  Ctrl+U / Ctrl+D       scroll one page (Warp-friendly)
  u / d / Space         page up / down
  ↑ ↓  j k              line scroll
  history mode          scroll up — new logs won't jump the view
  PgDown to bottom      resume live tail

Actions
  f                     filter logs by service
  q / Ctrl+C            quit (waits for clean shutdown)
  k                     free occupied port & restart
  a                     attach to process on conflict port
  n / Enter             ignore port conflict

Startup
  sync (default)        start services in dependency order, sequentially
  async                 launch services in parallel
  Set in muxdev.yaml: runtime: sync|async  or  muxdev --runtime async`,
			TryCommand: "muxdev help runtime-tui",
		},
		{
			ID:       "config-yaml",
			Category: "Config",
			Title:    "muxdev.yaml",
			Body: `Example manifest:

  name: My App
  subtitle: Local stack
  runtime: sync

  services:
    api:
      label: API
      command: npm run dev:api
      port: "${API_PORT}"
      env:
        NODE_ENV: development
      depends_on: []

    web:
      label: Web UI
      command: npm run dev:web
      port: "${UI_PORT}"
      depends_on: [api]

Ports use the same .env layering as listing: shell → .env.local → .env
→ .env.example → service env.

Commands run via sh -c from the directory containing muxdev.yaml.`,
			TryCommand: "muxdev list",
		},
		{
			ID:       "session-logs",
			Category: "Logs & sessions",
			Title:    "Session logs on disk",
			Body: `Every run writes a session under the platform sessions directory:

  Linux    $XDG_STATE_HOME/muxdev/sessions  (~/.local/state/…)
  macOS    ~/Library/Application Support/muxdev/sessions
  Windows  %LOCALAPPDATA%\muxdev\sessions

Each session folder contains:
  meta.json      project, services, runtime, timestamps
  session.log    [label] line output

Browse
  muxdev logs              interactive session browser
  muxdev logs list         table (this project)
  muxdev logs list --all   all projects
  muxdev logs show --latest
  muxdev logs path         print sessions directory`,
			TryCommand: "muxdev logs path",
		},
		{
			ID:       "port-conflicts",
			Category: "Troubleshooting",
			Title:    "Port conflicts",
			Body: `When a service can't bind its port, muxdev detects it from logs and
offers actions:

  k   stop processes on that port, then restart muxdev services
  a   attach to the existing process and stream its output
  n   ignore and continue

Tips
  • Log lines with an explicit port (EADDRINUSE :::4000) win over config hints
  • uvicorn [Errno 98] uses the service's resolved bind port from env
  • Quit with q so muxdev kills child process groups — avoids stale ports
  • Or: muxdev logs path  then inspect session.log after a crash`,
			TryCommand: "muxdev list",
		},
	}
}
