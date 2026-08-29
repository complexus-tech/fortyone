dev:
	@set -e; \
	pnpm dev & \
	pnpm_pid=$$!; \
	$(MAKE) -C apps/server dev & \
	server_pid=$$!; \
	trap 'kill $$pnpm_pid $$server_pid 2>/dev/null || true' INT TERM EXIT; \
	while kill -0 $$pnpm_pid 2>/dev/null && kill -0 $$server_pid 2>/dev/null; do \
		sleep 1; \
	done; \
	kill $$pnpm_pid $$server_pid 2>/dev/null || true; \
	wait $$pnpm_pid 2>/dev/null || pnpm_status=$$?; \
	wait $$server_pid 2>/dev/null || server_status=$$?; \
	if [ "$${pnpm_status:-0}" -ne 0 ] && [ "$${pnpm_status:-0}" -ne 130 ] && [ "$${pnpm_status:-0}" -ne 143 ]; then \
		exit $$pnpm_status; \
	fi; \
	if [ "$${server_status:-0}" -ne 0 ] && [ "$${server_status:-0}" -ne 130 ] && [ "$${server_status:-0}" -ne 143 ]; then \
		exit $$server_status; \
	fi

dev-projects:
	@set -e; \
	pnpm dev:projects & \
	pnpm_pid=$$!; \
	$(MAKE) -C apps/server dev & \
	server_pid=$$!; \
	trap 'kill $$pnpm_pid $$server_pid 2>/dev/null || true' INT TERM EXIT; \
	while kill -0 $$pnpm_pid 2>/dev/null && kill -0 $$server_pid 2>/dev/null; do \
		sleep 1; \
	done; \
	kill $$pnpm_pid $$server_pid 2>/dev/null || true; \
	wait $$pnpm_pid 2>/dev/null || pnpm_status=$$?; \
	wait $$server_pid 2>/dev/null || server_status=$$?; \
	if [ "$${pnpm_status:-0}" -ne 0 ] && [ "$${pnpm_status:-0}" -ne 130 ] && [ "$${pnpm_status:-0}" -ne 143 ]; then \
		exit $$pnpm_status; \
	fi; \
	if [ "$${server_status:-0}" -ne 0 ] && [ "$${server_status:-0}" -ne 130 ] && [ "$${server_status:-0}" -ne 143 ]; then \
		exit $$server_status; \
	fi

.PHONY: dev dev-projects
