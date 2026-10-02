# Load with: source <(vek bonsai init bash)
vek() {
    if [[ ${1-} == bonsai ]]; then
        shift
        bonsai "$@"
    else
        command vek "$@"
    fi
}

bonsai() {
    if [[ $# -eq 0 ]]; then
        command vek bonsai
        return
    fi

    local action=$1
    shift
    case "$action" in
        clone|start|checkout|clean)
            local destination
            destination="$(command vek bonsai "$action" --print-path "$@" && printf '.')" || return
            destination=${destination%.}
            destination=${destination%$'\n'}
            builtin cd -- "$destination"
            ;;
        *)
            command vek bonsai "$action" "$@"
            ;;
    esac
}
