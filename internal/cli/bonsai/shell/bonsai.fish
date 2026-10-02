# Load with: vek bonsai init fish | source
function vek --description 'Run vek and enter Bonsai worktrees'
    if test (count $argv) -gt 0; and test "$argv[1]" = bonsai
        bonsai $argv[2..-1]
    else
        command vek $argv
    end
end

function bonsai --description 'Manage Bonsai worktrees and enter new directories'
    if test (count $argv) -eq 0
        command vek bonsai
        return $status
    end

    set -l action $argv[1]
    set -l arguments $argv[2..-1]
    switch $action
        case clone start checkout clean
            set -l destination (command vek bonsai $action --print-path $arguments)
            or return $status
            builtin cd -- "$destination"
        case '*'
            command vek bonsai $action $arguments
    end
end
