/*  Construido como parte da disciplina: FPPD - PUCRS - Escola Politecnica
    Professor: Fernando Dotti  (https://fldotti.github.io/)
    Modulo representando Algoritmo de Exclusão Mútua Distribuída:
    Semestre 2023/1
	Aspectos a observar:
	   mapeamento de módulo para estrutura
	   inicializacao
	   semantica de concorrência: cada evento é atômico
	   							  módulo trata 1 por vez
	Q U E S T A O
	   Além de obviamente entender a estrutura ...
	   Implementar o núcleo do algoritmo ja descrito, ou seja, o corpo das
	   funcoes reativas a cada entrada possível:
	   			handleUponReqEntry()  // recebe do nivel de cima (app)
				handleUponReqExit()   // recebe do nivel de cima (app)
				handleUponDeliverRespOk(msgOutro)   // recebe do nivel de baixo
				handleUponDeliverReqEntry(msgOutro) // recebe do nivel de baixo
*/

package DIMEXcomSnapShot

import (
	PP2PLink "SD/PP2PLink"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ------------------------------------------------------------------------------------
// ------- principais tipos
// ------------------------------------------------------------------------------------

type State int // enumeracao dos estados possiveis de um processo
const (
	noMX State = iota // %GS: iota = contador que vai 0, 1, 2, ...
	wantMX
	inMX
)

type dmxReq int // enumeracao dos estados possiveis de um processo
const (
	ENTER dmxReq = iota
	EXIT
	SNAPSHOT // %GS: adicionando para pedir um snapshot
)

type dmxResp struct { // mensagem do módulo DIMEX infrmando que pode acessar - pode ser somente um sinal (vazio)
	// mensagem para aplicacao indicando que pode prosseguir
}

type DIMEX_Module struct {
	Req       chan dmxReq  // canal para receber pedidos da aplicacao (REQ e EXIT)
	Ind       chan dmxResp // canal para informar aplicacao que pode acessar
	addresses []string     // endereco de todos, na mesma ordem
	id        int          // identificador do processo - é o indice no array de enderecos acima
	st        State        // estado deste processo na exclusao mutua distribuida
	waiting   []bool       // processos aguardando tem flag true
	lcl       int          // relogio logico local
	reqTs     int          // timestamp local da ultima requisicao deste processo
	nbrResps  int
	dbg       bool

	Pp2plink *PP2PLink.PP2PLink // acesso aa comunicacao enviar por PP2PLinq.Req  e receber por PP2PLinq.Ind
}

type SnapShot_Module struct { // GS: como modelar isso:
	// um módulo paralelo que tem referência ao DIMEX?
	// uma "sobrecarga" do DIMEX?
	DIMEX       DIMEX_Module
	file        os.File
	received    []int // 0 = Nada, 1 = Enviou take snapshot e está esperando resposta, 2 = recebeu a resposta
	ls          *localState
	isRecording bool // %GS: flag para gravar mensagens
	dbg         bool
	// Mais algo?
}

type localState struct { // GS: Um tipo para conter as informações que vão ser salvas na snapshot (incompleto?)
	DIMEX_State State
	lcl         int
	reqTs       int
	nbrResps    int
	waiting     []bool
	channels    []string
}

// ------------------------------------------------------------------------------------
// ------- inicializacao
// ------------------------------------------------------------------------------------

func NewSnapShot_Module(_addresses []string, _id int, _dbg bool, _dbgDIMEX bool) *SnapShot_Module {

	p2p := PP2PLink.NewPP2PLink(_addresses[_id], _dbg)

	dmx := &DIMEX_Module{
		Req: make(chan dmxReq, 1),
		Ind: make(chan dmxResp, 1),

		addresses: _addresses,
		id:        _id,
		st:        noMX,
		waiting:   make([]bool, len(_addresses)),
		lcl:       0,
		reqTs:     0,
		dbg:       _dbgDIMEX,

		Pp2plink: p2p}

	for i := 0; i < len(dmx.waiting); i++ {
		dmx.waiting[i] = false
	}

	file, err := os.OpenFile(fmt.Sprintf("./snapshots_%d.txt", _id), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		fmt.Println("Error opening file:", err)
		panic("AAAAAAAAAAAAAAAAAA")
	}

	snp := &SnapShot_Module{
		DIMEX:       *dmx,
		file:        *file,
		ls:          nil,
		isRecording: false,
		received:    make([]int, len(_addresses)),
		dbg:         _dbg}

	for i := 0; i < len(snp.received); i++ {
		snp.received[i] = 0
	}
	fmt.Printf("len(dmx.addresses) = %d\n", len(dmx.addresses))
	snp.Start()
	snp.DIMEX.outDbg("Init DIMEX!")
	return snp
}

// ------------------------------------------------------------------------------------
// ------- nucleo do funcionamento
// ------------------------------------------------------------------------------------

func (module *SnapShot_Module) Start() {

	go func() {
		for {
			select {
			case dmxR := <-module.DIMEX.Req: // vindo da  aplicação
				{
					module.DIMEX.outDbg("dmxR recebeu module.Req")
					if dmxR == ENTER {
						module.DIMEX.outDbg("app pede mx")
						module.DIMEX.handleUponReqEntry() // ENTRADA DO ALGORITMO

					} else if dmxR == EXIT {
						module.DIMEX.outDbg("app libera mx")
						module.DIMEX.handleUponReqExit() // ENTRADA DO ALGORITMO
					} else if dmxR == SNAPSHOT {
						// %GS: devia fazer exatamente igual e mandar uma mensagem pra si mesmo, mas não sei se isso aqui funciona
						module.DIMEX.sendToLink(module.DIMEX.addresses[module.DIMEX.id], "takeSnapshot "+fmt.Sprint(module.DIMEX.id), "starting takeSnapshot")
					}
				}

			case msgOutro := <-module.DIMEX.Pp2plink.Ind: // vindo de outro processo
				{
					if module.isRecording {
						id_do_remetente, err := strconv.Atoi(strings.Split(msgOutro.Message, " ")[1])
						if err != nil {
							println(err)
						}
						if module.received[id_do_remetente] == 2 {
							break // %GS: não grava se já recebeu a resposta desse remetente
						}
						// %GS: Salvando mensagem da forma mais básica possível
						module.ls.channels[id_do_remetente] += msgOutro.Message + ";\n"
					}
					//fmt.Printf("dimex recebe da rede: %s", msgOutro)
					if strings.Contains(msgOutro.Message, "respOk") {
						module.DIMEX.outDbg("         <<<---- responde! " + msgOutro.Message)
						module.DIMEX.handleUponDeliverRespOk(msgOutro) // ENTRADA DO ALGORITMO

					} else if strings.Contains(msgOutro.Message, "reqEntry") {
						module.DIMEX.outDbg("          <<<---- pede??  " + msgOutro.Message)
						module.DIMEX.handleUponDeliverReqEntry(msgOutro) // ENTRADA DO ALGORITMO

					} else if strings.Contains(msgOutro.Message, "takeSnapshot") {
						module.outDbg("          <<<---- snapshot??  " + msgOutro.Message)
						module.handleUponDeliverTakeSnapshot(msgOutro) // %GS: Entrada do snapshot
						module.outDbg("Algo aconteceu")
					}
				}
			}
		}
	}()
}

// ------------------------------------------------------------------------------------
// @@@@@@@ SnapShot
// -------
// -------
// ------------------------------------------------------------------------------------

/*
Protocolo de Chandy-Lamport (retirado direto dos slides)

 1. Processo p0 manda mensagem para si mesmo com “take snapshot”

 2. Seja pf o processo do qual pi recebe mensagem “take snapshot” pela primeira vez.
    Ao receber, pi grava seu estado local σi e envia a mensagem “take snapshot” a todos
    canais de saída em OUTi O estado de xf,i é setado vazio. Pi inicia a gravação de
    mensagens recebidas de cada um de seus outros canais em INi

 3. Seja ps o processo do qual pi recebe mensagem “take snapshot” depois da primeira
    vez, pi pára de gravar mensagens de ps e declara o estado xs,i como sendo as mensagens gravadas.

    Quando o processo pi tiver recebido “take snapshot” em todos canais de entrada,
    sua contribuição para o snapshot acaba.
    Este acaba pois a mensagem é enviada somente uma vez em cada canal de saída.
*/
func (module *SnapShot_Module) handleUponDeliverTakeSnapshot(msgOutro PP2PLink.PP2PLink_Ind_Message) {
	id_do_outro, err := strconv.Atoi(strings.Split(msgOutro.Message, " ")[1])
	module.outDbg("Oi")
	if err != nil {
		println(err)
	}
	if module.received[id_do_outro] == 0 {
		module.outDbg("Iniciando uma snapshot")
		// É a primeira vez (pedido)
		module.saveLocalState()
		module.isRecording = true // %GS: passa a gravar mensagens nos canais
		for i := 0; i < len(module.DIMEX.addresses); i++ {
			module.ls.channels[i] = ""
			if i == module.DIMEX.id {
				module.received[i] = 2 // Como não tem canal entre um processo e ele mesmo, só pula o estado 1
				continue
			}
			module.outDbg("          requisita ---->>> " + msgOutro.Message)
			module.DIMEX.sendToLink(module.DIMEX.addresses[i], "takeSnapshot "+fmt.Sprint(module.DIMEX.id), "takeSnapshot from "+strconv.Itoa(module.DIMEX.id))
			module.received[i] = 1
		}
		module.received[id_do_outro] = 2 // %GS: quem iniciou não vai responder, então não espera a resposta dele
	} else if module.received[id_do_outro] == 1 {
		module.outDbg("Mensagem de resposta recebida")
		// Não é a primeira vez (resposta)
		module.received[id_do_outro] = 2
		module.checkSnapshotEnd()
	}
	// Se receber de um com estado 2, só ignora
}

func (module *SnapShot_Module) saveLocalState() {
	ls := &localState{
		DIMEX_State: module.DIMEX.st,
		lcl:         module.DIMEX.lcl,
		reqTs:       module.DIMEX.reqTs,
		nbrResps:    module.DIMEX.nbrResps,
		waiting:     module.DIMEX.waiting,
		channels:    make([]string, len(module.DIMEX.addresses))}

	module.ls = ls

}

func (module *SnapShot_Module) checkSnapshotEnd() {
	module.outDbg("Verificando se acabou")
	for i := 0; i < len(module.DIMEX.addresses); i++ {
		module.outDbg(fmt.Sprintf("%d", module.received[i]))
	}
	for i := 0; i < len(module.DIMEX.addresses); i++ {
		if module.received[i] != 2 {
			module.outDbg("Não acabou")
			return
		}
	}
	module.outDbg("Terminado uma snapshot, escrevendo...")
	// %GS: Se chegou aqui, então todos estão terminados
	module.writeSnapshot()
	module.outDbg("Escrito")
	// %GS: reinicia tudo e espera o próximo ciclo
	for i := 0; i < len(module.DIMEX.addresses); i++ {
		module.received[i] = 0
	}
	module.outDbg("Pronto pra próxima")
}

// %GS: Função só é chamada quando todos retornam, então se der postergação indefinida de um processo não vai ter snapshot...
func (module *SnapShot_Module) writeSnapshot() {
	module.file.WriteString(fmt.Sprintf("DIMEX_State: %d\n", module.DIMEX.st))
	module.file.WriteString(fmt.Sprintf("lcl: %d\n", module.DIMEX.lcl))
	module.file.WriteString(fmt.Sprintf("reqTs: %d\n", module.DIMEX.reqTs))
	module.file.WriteString(fmt.Sprintf("nbrResps: %d\n", module.DIMEX.nbrResps))
	module.file.WriteString("waiting: ")
	for i := 0; i < len(module.received); i++ {
		module.file.WriteString(fmt.Sprintf("%t ", module.ls.waiting[i]))
	}
	module.file.WriteString("\nmessages in channels: \n")
	for i := 0; i < len(module.received); i++ {
		module.file.WriteString(module.ls.channels[i])
	}
	module.file.WriteString("\n")
}

// ------------------------------------------------------------------------------------
// @@@@@@@ DIMEX
// ------- tratamento de pedidos vindos da aplicacao
// ------- UPON ENTRY
// ------- UPON EXIT
// ------------------------------------------------------------------------------------

func (module *DIMEX_Module) handleUponReqEntry() {
	/*
					upon event [ dmx, Entry  |  r ]  do
		    			lts.ts++
		    			myTs := lts
		    			resps := 0
		    			para todo processo p
							trigger [ pl , Send | [ reqEntry, r, myTs ]
		    			estado := queroSC
	*/
	module.lcl++
	module.reqTs = module.lcl
	module.nbrResps = 0
	for i := 0; i < len(module.addresses); i++ {
		if i == module.id {
			continue
		}
		module.sendToLink(module.addresses[i], "reqEntry "+fmt.Sprint(module.id)+" "+fmt.Sprint(module.lcl), "space")
	}
	module.st = wantMX

}

func (module *DIMEX_Module) handleUponReqExit() {
	/*
						upon event [ dmx, Exit  |  r  ]  do
		       				para todo [p, r, ts ] em waiting
		          				trigger [ pl, Send | p , [ respOk, r ]  ]
		    				estado := naoQueroSC
							waiting := {}
	*/
	for i := 0; i < len(module.waiting); i++ {
		if module.waiting[i] {
			module.sendToLink(module.addresses[i], "respOk "+fmt.Sprint(module.id), "respOk (reqExit) from "+strconv.Itoa(module.id))
		}
	}
	module.st = noMX
	module.waiting = make([]bool, len(module.addresses))
	module.outDbg(fmt.Sprintf("module.waiting = %v\n", module.waiting))
}

// ------------------------------------------------------------------------------------
// ------- tratamento de mensagens de outros processos
// ------- UPON respOK
// ------- UPON reqEntry
// ------------------------------------------------------------------------------------

// %GS: chamado quando recebe um RespOk
func (module *DIMEX_Module) handleUponDeliverRespOk(msgOutro PP2PLink.PP2PLink_Ind_Message) {
	/*
						upon event [ pl, Deliver | p, [ respOk, r ] ]
		      				resps++
		      				se resps = N
		    				então trigger [ dmx, Deliver | free2Access ]
		  					    estado := estouNaSC

	*/
	module.nbrResps++
	module.outDbg(fmt.Sprintf("%d tem %d oks\n", module.id, module.nbrResps))
	if module.nbrResps == len(module.addresses)-1 {
		module.outDbg(fmt.Sprintf("%d pode entrar na seção crítica\n", module.id))
		module.Ind <- dmxResp{}
		module.st = inMX
	}

}

// %GS: chamado quando recebe um ReqEntry
func (module *DIMEX_Module) handleUponDeliverReqEntry(msgOutro PP2PLink.PP2PLink_Ind_Message) {
	// outro processo quer entrar na SC
	/*
						upon event [ pl, Deliver | p, [ reqEntry, r, rts ]  do
		     				se (estado == naoQueroSC)   OR
		        				 (estado == QueroSC AND  myTs >  ts)
							então  trigger [ pl, Send | p , [ respOk, r ]  ]
		 					senão
		        				se (estado == estouNaSC) OR
		           					 (estado == QueroSC AND  myTs < ts)
		        				então  postergados := postergados + [p, r ]
		     					lts.ts := max(lts.ts, rts.ts)
	*/
	// %GS: da onde vem o ID e timestamp do outro?
	id_do_outro, err := strconv.Atoi(strings.Split(msgOutro.Message, " ")[1])
	lcl_do_outro, err := strconv.Atoi(strings.Split(msgOutro.Message, " ")[2])
	if err != nil {
		println(err)
	}
	// println()
	// println(lcl_do_outro)
	// println(id_do_outro)
	if module.st == noMX ||
		(module.st == wantMX && before(id_do_outro, lcl_do_outro, module.id, module.reqTs)) {
		module.outDbg(fmt.Sprintf("%d deixou %d passar na frente\n", module.id, id_do_outro))
		module.sendToLink(module.addresses[id_do_outro], "respOk "+fmt.Sprint(module.id), "respOk (reqEntry) from "+strconv.Itoa(module.id))
	} else {
		module.outDbg(fmt.Sprintf("%d NÃO deixou %d passar na frente\n", module.id, id_do_outro))
		// %GS: Quase certo que não é necessário esse if mas como está no algoritmo fica por enquanto
		if module.st == inMX || (module.st == wantMX && before(module.id, module.reqTs, id_do_outro, lcl_do_outro)) {
			module.waiting[id_do_outro] = true
		}
		// %GS: tem que ter uma maneira mais bonita de fazer isso né...
		//module.lcl = int(math.Max(float64(module.lcl), float64(lcl_do_outro)))

		// %GS: reescrevendo a mesma coisa mas sem as conversões pra float e int
		if module.lcl > lcl_do_outro {
			module.lcl = lcl_do_outro
		}
	}
}

// ------------------------------------------------------------------------------------
// ------- funcoes de ajuda
// ------------------------------------------------------------------------------------

func (module *DIMEX_Module) sendToLink(address string, content string, space string) {
	module.outDbg(space + " ---->>>>   to: " + address + "     msg: " + content)
	module.Pp2plink.Req <- PP2PLink.PP2PLink_Req_Message{
		To:      address,
		Message: content}
}

// %GS: usar em caso de ter duas (ou mais) requisições simultâneas para decidir quem tem prioridade
func before(oneId, oneTs, othId, othTs int) bool {
	if oneTs < othTs {
		return true
	} else if oneTs > othTs {
		return false
	} else {
		return oneId < othId
	}
}

func (module *DIMEX_Module) outDbg(s string) {
	if module.dbg {
		fmt.Println(". . . . . . . . . . . . [ DIMEX : " + s + " ]")
	}
}
func (module *SnapShot_Module) outDbg(s string) {
	if module.dbg {
		fmt.Println(". . . . . . . . . . . . [ SNAPSHOT : " + s + " ]")
	}
}
