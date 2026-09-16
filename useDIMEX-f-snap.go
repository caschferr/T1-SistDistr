// Construido como parte da disciplina: Sistemas Distribuidos - PUCRS - Escola Politecnica
//  Professor: Fernando Dotti  (https://fldotti.github.io/)
// Uso p exemplo:
//   go run useDIMEX-f.go 0 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002
//   go run useDIMEX-f.go 1 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002
//   go run useDIMEX-f.go 2 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002
// ----------
// LANCAR N PROCESSOS EM SHELL's DIFERENTES, UMA PARA CADA PROCESSO.
// para cada processo fornecer: seu id único (0, 1, 2 ...) e a mesma lista de processos.
// o endereco de cada processo é o dado na lista, na posicao do seu id.
// no exemplo acima o processo com id=1  usa a porta 6001 para receber e as portas
// 5000 e 7002 para mandar mensagens respectivamente para processos com id=0 e 2
// -----------
// Esta versão supõe que todos processos tem acesso a um mesmo arquivo chamado "mxOUT.txt"
// Todos processos escrevem neste arquivo, usando o protocolo dimex para exclusao mutua.
// Os processos escrevem "|." cada vez que acessam o arquivo.   Assim, o arquivo com conteúdo
// correto deverá ser uma sequencia de
// |.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.
// |.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.
// |.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.
// etc etc ...     ....  até o usuário interromper os processos (ctl c).
// Qualquer padrao diferente disso, revela um erro.
//      |.|.|.|.|.||..|.|.|.  etc etc  por exemplo.
// Se voce retirar o protocolo dimex vai ver que o arquivo poderá entrelacar
// "|."  dos processos de diversas diferentes formas.
// Ou seja, o padrão correto acima é garantido pelo dimex.
// Ainda assim, isto é apenas um teste.  E testes são frágeis em sistemas distribuídos.

package main

import (
	"SD/DIMEXcomSnapShot"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Please specify at least one address:port!")
		fmt.Println("go run usaDIMEX-f-snap.go 0 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
		fmt.Println("go run usaDIMEX-f-snap.go 1 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
		fmt.Println("go run usaDIMEX-f-snap.go 2 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
		return
	}

	id, _ := strconv.Atoi(os.Args[1])
	addresses := os.Args[2:]
	// fmt.Print("id: ", id, "   ") fmt.Println(addresses)

	var dmx *DIMEXcomSnapShot.SnapShot_Module = DIMEXcomSnapShot.NewSnapShot_Module(addresses, id, true, false)
	fmt.Println(dmx)

	// abre arquivo que TODOS processos devem poder usar
	file, err := os.OpenFile("./mxOUT.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	logFile, err := os.OpenFile("./log.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		fmt.Println("Error opening logFile:", err)
		return
	}
	defer logFile.Close()
	defer file.Close() // Ensure the file is closed at the end of the function

	// espera para facilitar inicializacao de todos processos (a mao)
	time.Sleep(3 * time.Second)

	go func() {
		// %GS: Tira um snapshot a cada 5 segundos, eu acho
		if id == 0 {
			for i := 0; i == i; {
				time.Sleep(5 * time.Second)
				dmx.DIMEX.Req <- DIMEXcomSnapShot.SNAPSHOT
			}
		}
	}()

	for {
		// SOLICITA ACESSO AO DIMEX
		fmt.Println("[ APP id: ", id, " PEDE   MX ]")
		dmx.DIMEX.Req <- DIMEXcomSnapShot.ENTER
		//fmt.Println("[ APP id: ", id, " ESPERA MX ]")
		// ESPERA LIBERACAO DO MODULO DIMEX
		<-dmx.DIMEX.Ind //

		// A PARTIR DAQUI ESTA ACESSANDO O ARQUIVO SOZINHO
		// _, err = logFile.WriteString(strconv.Itoa(id) + " ")

		// _, err := file.Seek(0, io.SeekEnd)
		// if err != nil {
		// 	panic("deu merda " + err.Error())
		// }
		// buf := make([]byte, 1)
		// _, err = file.Read(buf)
		// if err != nil && err != io.EOF {
		// 	panic("deu merda 2 o retorno " + err.Error())
		// }
		// if string(buf[0]) == "|" {
		// 	_, err = logFile.WriteString("\nREPEATED | FROM ID " + strconv.Itoa(id))
		// 	if err != nil {
		// 		fmt.Println("Could not write to logFile")
		// 	}
		// 	panic("REPEATED OUTPUT |")
		// }

		_, err = file.WriteString("|") // marca entrada no arquivo
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}

		fmt.Println("[ APP id: ", id, " *EM*   MX ]")

		// _, err = file.Seek(-1, io.SeekEnd)
		// if err != nil {
		// 	panic("deu merda " + err.Error())
		// }
		// buf = make([]byte, 1)
		// _, err = file.Read(buf)
		// if err != nil && err != io.EOF {
		// 	panic("deu merda 2 o retorno " + err.Error())
		// }
		// if string(buf[0]) == "." {
		// 	_, err = logFile.WriteString("\nREPEATED . FROM ID " + strconv.Itoa(id))
		// 	if err != nil {
		// 		fmt.Println("Could not write to logFile")
		// 	}
		// 	panic("REPEATED OUTPUT .")
		// }
		_, err = file.WriteString(".") // marca saida no arquivo
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}

		// AGORA VAI LIBERAR O ARQUIVO PARA OUTROS
		dmx.DIMEX.Req <- DIMEXcomSnapShot.EXIT //
		fmt.Println("[ APP id: ", id, " FORA   MX ]")
	}
}
