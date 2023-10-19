// Código exemplo para o trabaho de sistemas distribuidos (eleicao em anel)
// By Cesar De Rose - 2022

package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmacao da eleicao)
	corpo [3]int // conteudo da mensagem para colocar os ids (usar um tamanho ocmpativel com o numero de processos no anel)
	//corpo só existe se eleição está sendo feita
}

var (
	chans = []chan mensagem{ // vetor de canias para formar o anel de eleicao - chan[0], chan[1] and chan[2] ...
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	controle = make(chan int)
	wg       sync.WaitGroup // wg is used to wait for the program to finish
)

func ElectionControler(in chan int) {
	defer wg.Done()

	//fazer processos falharem

	//fazer processos voltarem a funcionar

	var temp mensagem

	// comandos para o anel iniciam aqui

	// mudar o processo 0 - canal de entrada 3 - para falho (defini mensagem tipo 2 pra isto)

	temp.tipo = 2
	chans[3] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")

	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// mudar o processo 1 - canal de entrada 0 - para falho (defini mensagem tipo 2 pra isto)

	temp.tipo = 2
	chans[0] <- temp
	fmt.Printf("Controle: mudar o processo 1 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// matar os outros processos com mensagens não conhecidas (só pra cosumir a leitura)

	temp.tipo = 4
	chans[1] <- temp
	chans[2] <- temp

	fmt.Println("\n   Processo controlador concluído\n")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()

	// variaveis locais que indicam se este processo é o lider e se esta ativo

	var actualLeader int
	var bFailed bool = false // todos iniciam sem falha

	actualLeader = leader // indicação do lider veio por parâmatro

	temp := <-in // ler mensagem
	fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])

	//fazer votação para novo líder quando processo falha
	//colocar seu próprio id na mensagem

	//processo que iniciou a votação recebe a mensagem
	//o escolhido é aquele que tem o maior id

	//dispara eleição
	//cada processo espera por uma mensagem e quando líder falha ele envia mensagem
	//primeiro processo a consumir a mensagem é o processo que dispara a nova eleição

	switch temp.tipo {
	case 1: //processo líder falha
		{
			bFailed = true
			//processo líder manda mensagem pro próximo avisando que falhou
			var m mensagem
			m.tipo = 2 //iniciar eleicao
			out <- m
		}
	case 2: //quem dispara a eleição -> disparando a eleição
		{
			var novaMensagem mensagem
			novaMensagem.tipo = 3

			for i := 0; i < 3; i++ {
				novaMensagem.corpo[i] = -1
			}

			novaMensagem.corpo[0] = TaskId
			out <- novaMensagem

			// bFailed = true
			// fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
			// fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
			// controle <- -5
		}
	case 3: //fazendo uma eleição -> votação
		{
			//colocar id próprio no corpo
			//passar corpo e tipo pro próximo

			m1 := <-in
			if bFailed == false {
				cont := 0

				for i := 0; i < 3; i++ {
					if m1.corpo[i] == -1 {
						m1.corpo[i] = TaskId
						cont++
					}
				}

				if cont <= 0 {
					m1.tipo = 5
				}

			}

			out <- m1

			// bFailed = false
			// fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
			// fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
			// controle <- -5
		}
	case 4: //avisando os processos quem ganhou a eleição
		{
			m1 := <-in
			actualLeader = m1.corpo[0]

			if m1.corpo[2] == m1.corpo[1] {
				//parar de mandar mensagem desse tipo
				controle <- -5
			} else if m1.corpo[1] == m1.corpo[0] {
				m1.corpo[2] = actualLeader
				out <- m1
			} else {
				m1.corpo[1] = actualLeader
				out <- m1
			}

			//fazer votação para novo líder quando processo falha
			//colocar seu próprio id na mensagem
			//processo que falhou não pode enviar nem receber mensagem

			// fmt.Printf()

		}
	case 5: //decide quem ganha a eleição e começa a avisar quem ganhou a eleição
		{
			ganhador := -1
			m1 := <-in
			for i := 0; i < 3; i++ {
				if m1.corpo[i] > ganhador {
					ganhador = m1.corpo[i]
				}
			}

			actualLeader = ganhador

			m1.corpo[0] = actualLeader
			m1.tipo = 4

			out <- m1

			//processo que iniciou a votação recebe a mensagem
			//o escolhido é aquele que tem o maior id

			//mudar actualLeader?
		}
	case 6: //fazer processo voltar
		{
			bFailed = false
		}
	case 7:
		{
			//faz todo o processo de disparar eleição até saber quem ganhou
		}
	default:
		{
			fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
			fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
		}
	}

	fmt.Printf("%2d: terminei \n", TaskId)
}

func main() {

	wg.Add(5) // Add a count of four, one for each goroutine

	// criar os processo do anel de eleicao

	go ElectionStage(0, chans[3], chans[0], 0) // este é o lider
	go ElectionStage(1, chans[0], chans[1], 0) // não é lider, é o processo 0
	go ElectionStage(2, chans[1], chans[2], 0) // não é lider, é o processo 0
	go ElectionStage(3, chans[2], chans[3], 0) // não é lider, é o processo 0

	fmt.Println("\n   Anel de processos criado")

	// criar o processo controlador

	go ElectionControler(controle)

	fmt.Println("\n   Processo controlador criado\n")

	wg.Wait() // Wait for the goroutines to finish\
}
