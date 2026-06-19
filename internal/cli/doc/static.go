package doc

// staticPages returns curated wiki topics that are not tied to a single subcommand.
func staticPages() []Page {
	return []Page{
		{
			ID:       "welcome",
			Category: "Start here",
			Title:    "Welcome",
			Summary:  "Run your local dev stack from muxdev.yaml with an interactive TUI.",
			Body: `muxdev orchestrates local services defined in muxdev.yaml — pick what to run,
multiplex logs, resolve port conflicts, and review past sessions.

Install (public GitHub Releases):
  curl -fsSL https://raw.githubusercontent.com/yarkingulacti/muxdev-cli/master/scripts/install.sh | bash

Quick start:
  1. cd your-project
  2. muxdev init
  3. muxdev

Update:
  muxdev update --check
  muxdev update --yes

No config yet? Starting muxdev in a new repo offers the init wizard automatically.

Global flags (work with any command):
  -c, --config PATH       path to muxdev.yaml
      --no-interactive    plain stdout/stderr, no TUI
      --runtime MODE      sync (default) or async startup
      --focus IDS         comma-separated service ids

Common next steps:
  muxdev list             table of services and resolved ports
  muxdev configure        edit the manifest interactively
  muxdev logs             browse saved runtime sessions

Tip: Press t on any topic to run its example command safely.`,
			TryCommand: "muxdev version",
		},
		{
			ID:       "runtime-tui",
			Category: "Runtime",
			Title:    "Runtime TUI & keyboard",
			Summary:  "Keyboard shortcuts while services run and logs stream live.",
			Body: `While services run, muxdev shows a live log panel with history scroll.

Navigation:
  PgUp / PgDown         scroll one line
  Ctrl+U / Ctrl+D       scroll one page (Warp-friendly)
  u / d / Space         page up / down
  ↑ ↓  j k              line scroll
  scroll up             enter history mode — new logs won't jump the view
  PgDown to bottom      resume live tail

Actions:
  f                     filter logs by service
  r                     re-run services (interactive picker)
  ctrl+q / Ctrl+C       quit gracefully (wait for shutdown)
  ctrl+q again          force quit if shutdown is taking too long
  q                     force quit immediately
  k                     free occupied port and restart services
  a                     attach to process on conflict port
  n / Enter             ignore port conflict

Startup modes:
  sync (default)        start services in dependency order
  async                 launch services in parallel

Set runtime in muxdev.yaml or pass  muxdev --runtime async`,
			TryCommand: "muxdev help runtime-tui",
		},
		{
			ID:       "config-yaml",
			Category: "Config",
			Title:    "muxdev.yaml",
			Summary:  "Manifest format: services, ports, env, and dependencies.",
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

Port resolution:
  Shell env → .env.local → .env → .env.example → service env block

Commands run via sh -c from the directory containing muxdev.yaml.

Tip: Use muxdev configure to edit without touching YAML by hand.`,
			TryCommand: "muxdev list",
		},
		{
			ID:       "session-logs",
			Category: "Logs & sessions",
			Title:    "Session logs on disk",
			Summary:  "Every run writes a timestamped session you can browse later.",
			Body: `Each muxdev run saves output under the platform sessions directory:

  Linux    $XDG_STATE_HOME/muxdev/sessions
  macOS    ~/Library/Application Support/muxdev/sessions
  Windows  %LOCALAPPDATA%\muxdev\sessions

Session folder contents:
  meta.json      project, services, runtime, timestamps
  session.log    [label] prefixed log lines

Browse sessions:
  muxdev logs              interactive session browser
  muxdev logs list         table for this project
  muxdev logs list --all   all projects on this machine
  muxdev logs show --latest
  muxdev logs path         print sessions directory`,
			TryCommand: "muxdev logs path",
		},
		{
			ID:       "port-conflicts",
			Category: "Troubleshooting",
			Title:    "Port conflicts",
			Summary:  "What to do when a service cannot bind its port.",
			Body: `When a service can't bind its port, muxdev detects it from logs and offers:

  k   stop processes on that port, then restart muxdev services
  a   attach to the existing process and stream its output
  n   ignore and continue

Tips:
  • Log lines with an explicit port (EADDRINUSE :::4000) win over config hints
  • uvicorn [Errno 98] uses the service's resolved bind port from env
  • Quit with q so muxdev kills child process groups — avoids stale ports
  • Inspect session.log after a crash: muxdev logs path

Note: muxdev list shows resolved ports and env sources before you start.`,
			TryCommand: "muxdev list",
		},
	}
}
