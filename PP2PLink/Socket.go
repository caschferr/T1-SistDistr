package PP2PLink

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func Envia(lk *PP2PLink) { //Funciona como o Cliente, extrai o To e envia para a rede
	for requisicao := range lk.Req { //pega a requisição e separa em destino e mensagem
		destino := requisicao.To
		mensagem := requisicao.Message
		fmt.Println("Enviando mensagem: ", mensagem, " para ", destino)
		conexao, erro := net.Dial("tcp", destino)
		if erro != nil { //caso haja erro na conexão, imprime e pula para a próxima requisição
			fmt.Println("Erro ao conectar em ", destino, " : ", erro)
			continue
		}
		conexao.Write([]byte(mensagem + "\n"))
		conexao.Close()
	}
}

func Recebe(lk *PP2PLink, serverPort string) { //Funciona como o Servidor
	listener, erro := net.Listen("tcp", serverPort)
	if erro != nil {
		fmt.Println("Erro de recepção: ", erro)
	}
	for {
		 conexao, erro := listener.Accept()
		 if erro != nil {
		  fmt.Println("Erro ao estabelecer conexão: ", erro)
		 }
		 mensagem, erroR :=  bufio.NewReader(conexao).ReadString('\n')// Le os bytes da conexão, transforma em string até achar \n
		 if erroR != nil {
		  fmt.Println("Erro ao ler a mensagem: ", erroR)
		  conexao.Close()
		  continue
		 }
		 mensagemN := strings.TrimSpace(mensagem) //tira o /n
	     mensagemStruct := PP2PLink_Ind_Message{
          Message:  mensagemN,
		 }
		 lk.Ind <- mensagemStruct// Jogar essa mensagem lida no canal de entrada para avisar a aplicaçãos
		 conexao.Close()

	}

}
