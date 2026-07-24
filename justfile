# Build orchestration for plugin submodules.

root := justfile_directory()
main_output_dir := root / "bin"
output_dir := root / "plugins/bin"
version := env("VERSION", "devel")

# List available recipes.
default:
    @just --list

# Build monokit2 and every plugin submodule for the host platform.
build: build-main build-plugins

# Build the main monokit2 binary for the host platform.
build-main:
    @echo "Building monokit2 {{ version }} for the host platform"
    mkdir -p "{{ main_output_dir }}"
    rm -f "{{ main_output_dir }}/monokit2"
    go build -ldflags "-X main.version={{ version }}" -o "{{ main_output_dir }}/monokit2" ./main.go

# Build every plugin submodule and collect its host binary in plugins/bin.
build-plugins:
    #!/usr/bin/env bash
    set -euo pipefail

    root={{ quote(root) }}
    mapfile -t modules < <(
        git -C "$root" config --file .gitmodules \
            --get-regexp '^submodule\..*\.path$' |
            awk '$2 ~ /^plugins\// { print $2 }'
    )

    if (( ${#modules[@]} == 0 )); then
        echo "No plugin submodules are declared in .gitmodules" >&2
        exit 1
    fi

    for module in "${modules[@]}"; do
        plugin="${module##*/}"
        just --justfile "$root/justfile" --working-directory "$root" \
            build-plugin "$plugin"
    done

# Build one declared plugin submodule and move its binary to plugins/bin.
build-plugin plugin:
    #!/usr/bin/env bash
    set -euo pipefail

    root={{ quote(root) }}
    output_dir={{ quote(output_dir) }}
    plugin={{ quote(plugin) }}
    module="plugins/$plugin"
    module_dir="$root/$module"

    if ! git -C "$root" config --file .gitmodules \
        --get-regexp '^submodule\..*\.path$' |
        awk '{ print $2 }' |
        grep -Fxq "$module"; then
        echo "Plugin is not declared as a submodule: $module" >&2
        exit 1
    fi

    if [[ ! -e "$module_dir/.git" ]]; then
        echo "Plugin submodule is not initialized: $module" >&2
        echo "Run: git submodule update --init --recursive" >&2
        exit 1
    fi

    plugin_justfile="$module_dir/justfile"
    if [[ ! -f "$plugin_justfile" ]]; then
        echo "Plugin has no justfile: $module" >&2
        exit 1
    fi

    just --justfile "$plugin_justfile" --working-directory "$module_dir" build

    binary="$module_dir/bin/$plugin"
    if [[ ! -f "$binary" ]]; then
        echo "Plugin build did not produce the expected binary: $binary" >&2
        exit 1
    fi

    mkdir -p "$output_dir"
    mv -f -- "$binary" "$output_dir/$plugin"
    echo "Installed $plugin to $output_dir/$plugin"

# Fetch and check out the latest remote commit for every submodule.
sync-submodules:
    git submodule sync --recursive
    git submodule update --init --recursive --remote
    git submodule status --recursive
