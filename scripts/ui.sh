#!/bin/sh
# Presentation helpers for the Makefile. Cosmetic only: every function returns
# the wrapped command's exit status, and color/animation is disabled when
# stdout is not a terminal, NO_COLOR is set, or TERM=dumb.

set -u

color=0
if [ -n "${FORCE_COLOR-}" ]; then
	color=1
elif [ -t 1 ] && [ -z "${NO_COLOR-}" ] && [ "${TERM-}" != "dumb" ]; then
	color=1
fi

if [ "$color" -eq 1 ]; then
	bold=$(printf '\033[1m')
	dim=$(printf '\033[2m')
	reset=$(printf '\033[0m')
	cyan=$(printf '\033[36m')
	magenta=$(printf '\033[35m')
	green=$(printf '\033[32m')
	red=$(printf '\033[31m')
else
	bold='' dim='' reset='' cyan='' magenta='' green='' red=''
fi

now_ns() {
	stamp=$(date +%s%N 2>/dev/null || true)
	case $stamp in
	'' | *[!0-9]*) printf '%s' "$(($(date +%s) * 1000000000))" ;;
	*) printf '%s' "$stamp" ;;
	esac
}

elapsed_s() {
	finish=$(now_ns)
	delta=$((finish - $1))
	if [ "$delta" -ge 1000000000 ]; then
		printf '%d.%02ds' $((delta / 1000000000)) $(((delta % 1000000000) / 10000000))
	else
		printf '0.%02ds' $((delta / 10000000))
	fi
}

dashes() {
	count=$1
	while [ "$count" -gt 0 ]; do
		printf '─'
		count=$((count - 1))
	done
}

banner() {
	title=$(printf '%s' "${1-}" | tr '[:lower:]' '[:upper:]')
	width=47

	if [ "$color" -eq 0 ]; then
		printf '\n== %s ==\n' "$title"
		return 0
	fi

	pad=$((36 - ${#title}))
	[ "$pad" -lt 1 ] && pad=1
	spaces=$(printf '%*s' "$pad" '')

	printf '\n  %s╭%s' "$cyan" "$(dashes 24)"
	sleep 0.03
	printf '%s╮%s\n' "$(dashes 23)" "$reset"
	printf '  %s│%s %staskgo%s %s·%s %s%s%s%s %s│%s\n' \
		"$cyan" "$reset" "$bold" "$reset" "$magenta" "$reset" \
		"$bold" "$title" "$reset" "$spaces" "$cyan" "$reset"
	printf '  %s╰%s╯%s\n' "$cyan" "$(dashes "$width")" "$reset"
	return 0
}

ok() {
	if [ "$color" -eq 1 ]; then
		printf '  %s✔%s %s\n' "$green" "$reset" "${1-}"
	else
		printf '  %s\n' "${1-}"
	fi
}

fail() {
	if [ "$color" -eq 1 ]; then
		printf '  %s✖%s %s\n' "$red" "$reset" "${1-}"
	else
		printf '  %s FAILED\n' "${1-}"
	fi
}

# spin <label> -- <command...>
# Captures output so it cannot garble the spinner; prints it only on failure.
spin() {
	label=${1-}
	shift || true
	[ "${1-}" = "--" ] && shift

	tmp=$(mktemp 2>/dev/null || printf '/tmp/taskgo-ui.%s' "$$")
	start=$(now_ns)
	spinner=''

	if [ "$color" -eq 1 ]; then
		printf '  %s▸%s %s ' "$cyan" "$reset" "$label"
		(
			while :; do
				for frame in '⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏'; do
					printf '\b%s' "$frame"
					sleep 0.07
				done
			done
		) &
		spinner=$!
	fi

	"$@" >"$tmp" 2>&1
	status=$?

	if [ -n "$spinner" ]; then
		kill "$spinner" 2>/dev/null
		wait "$spinner" 2>/dev/null
		printf '\r\033[K'
	fi

	took=$(elapsed_s "$start")
	if [ "$status" -eq 0 ]; then
		if [ "$color" -eq 1 ]; then
			printf '  %s✔%s %s %s%s%s\n' "$green" "$reset" "$label" "$dim" "$took" "$reset"
		else
			printf '  %s ... ok (%s)\n' "$label" "$took"
		fi
	else
		fail "$label ($took)"
		sed 's/^/    /' "$tmp" 2>/dev/null
	fi
	rm -f "$tmp"
	return "$status"
}

# run <label> -- <command...>
# Streams output live, then prints the result line.
run() {
	label=${1-}
	shift || true
	[ "${1-}" = "--" ] && shift

	start=$(now_ns)
	if [ "$color" -eq 1 ]; then
		printf '  %s▸%s %s\n' "$cyan" "$reset" "$label"
	else
		printf '  %s\n' "$label"
	fi

	"$@"
	status=$?
	took=$(elapsed_s "$start")
	if [ "$status" -eq 0 ]; then
		ok "$label ($took)"
	else
		fail "$label ($took)"
	fi
	return "$status"
}

command=${1-}
[ $# -gt 0 ] && shift

case $command in
banner) banner "$@" ;;
ok) ok "$@" ;;
fail) fail "$@" ;;
spin) spin "$@" ;;
run) run "$@" ;;
*)
	printf 'usage: ui.sh banner|ok|fail|spin|run ...\n' >&2
	exit 2
	;;
esac
