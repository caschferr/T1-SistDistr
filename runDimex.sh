ARQUIVO="useDIMEX-f.go"
ARQUIVOSNAP="useDIMEX-f-snap.go"
ENDERECOS="127.0.0.1:5000 127.0.0.1:6001 127.0.0.1:7002"
DIR_ATUAL="$(pwd)"

rm mxOUT.txt
rm log.txt
rm snapshots_0.txt
rm snapshots_1.txt
rm snapshots_2.txt

abrir_terminal() {
    local id="$2"
    if [ "$1" -eq "1" ]; then
        local cmd="cd '$DIR_ATUAL' && go run $ARQUIVOSNAP $id $ENDERECOS; exec bash"
    else
        local cmd="cd '$DIR_ATUAL' && go run $ARQUIVO $id $ENDERECOS; exec bash"
    fi

    if command -v gnome-terminal >/dev/null 2>&1; then
        gnome-terminal -- bash -c "$cmd"
    elif command -v konsole >/dev/null 2>&1; then
        konsole -e bash -c "$cmd"
    elif command -v xfce4-terminal >/dev/null 2>&1; then
        xfce4-terminal -e "bash -c \"$cmd\""
    elif command -v xterm >/dev/null 2>&1; then
        xterm -e bash -c "$cmd" &
    elif command -v alacritty >/dev/null 2>&1; then
        alacritty -e bash -c "$cmd" &
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        osascript -e "tell application \"Terminal\" to do script \"$cmd\""
    else
        echo "Nenhum emulador de terminal suportado foi encontrado."
        echo "Rode manualmente: go run $ARQUIVO $id $ENDERECOS"
    fi
}

for id in 0 1 2; do
    abrir_terminal "$1" "$id"
    sleep 0.3
done