// Código exemplo para o trabaho de sistemas distribuidos (eleicao em anel)
// By Cesar De Rose - 2022

package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmacao da eleicao)
	corpo [3]int // conteudo da mensagem para colocar os ids (usar um tamanho compativel com o numero de processos no anel)
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
	wait     = make(chan int)
)

func ElectionControler(in chan int) {
	defer wg.Done()

	var temp mensagem

	// mudar o processo 0 - canal de entrada 3 - para falho (defini mensagem tipo 1 pra isto)
	temp.tipo = 1
	chans[2] <- temp
	fmt.Printf("Controle: mudar o processo 3 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// fazendo eleição pra novo líder
	temp.tipo = 2
	chans[3] <- temp
	fmt.Printf("Controle: processo 1 dispara eleição, eleição é feita e concluída, processos são avisados\n")
	actualLeader := <-in
	fmt.Printf("Controle: confirmação %d\n", actualLeader) // receber e imprimir confirmação

	// voltando o processo 3
	temp.tipo = 6
	temp.corpo[0] = actualLeader
	chans[2] <- temp
	fmt.Printf("Controle: mudar o processo 3 para funcional\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// temp.tipo = 2
	// chans[3] <- temp
	// fmt.Printf("Controle: processo 1 dispara eleição, eleição é feita e concluída, processos são avisados\n")
	// fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// mudar o processo 1 - canal de entrada 0 - para falho (defini mensagem tipo 2 pra isto)

	// temp.tipo = 2
	// chans[0] <- temp
	// fmt.Printf("Controle: mudar o processo 1 para falho\n")
	// fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// matar os outros processos
	temp.tipo = 7
	chans[0] <- temp
	chans[1] <- temp
	chans[2] <- temp
	chans[3] <- temp

	fmt.Println("\n   Processo controlador concluído\n")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()

	// variaveis locais que indicam se este processo é o lider e se esta ativo
	var actualLeader int
	var bFailed bool = false // todos iniciam sem falha

	actualLeader = leader // indicação do lider veio por parâmatro

	// temp := <-in // ler mensagem
	// fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])

	var finish bool = false

	for {
		select {
		case temp := <-in:
			fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])

			switch temp.tipo {
			case 1: //processo líder falha
				{
					bFailed = true
					fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
					fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
					controle <- -5
				}
			case 2: //quem dispara a eleição -> disparando a eleição
				{
					fmt.Printf("%2d: disparando a eleição\n", TaskId)

					temp.tipo = 3
					for i := 0; i < 3; i++ {
						temp.corpo[i] = -1
					}

					temp.corpo[0] = TaskId
					out <- temp
				}
			case 3: //fazendo uma eleição -> votação
				{
					// fmt.Printf("CASE 3 - %2d: recebi mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])

					if bFailed == false {
						for i := 0; i < 3; i++ {
							if temp.corpo[i] == -1 {
								temp.corpo[i] = TaskId
								break
							}
						}
					}

					// fmt.Printf("CASE 3 POS ARRAY PREENCHIDA - %2d: envia mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])

					var cont int
					cont = 0
					for i := 0; i < 3; i++ {
						if temp.corpo[i] != -1 {
							cont++
						}
					}

					if cont == 3 {
						temp.tipo = 5
					}

					fmt.Printf("fazendo uma eleição - %2d: envia mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])

					out <- temp

					// fmt.Printf("CASE 3 FINAL REAL - %2d: envia mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])
				}
			case 4: //avisando os processos quem ganhou a eleição
				{
					actualLeader = temp.corpo[0]

					if temp.corpo[2] == temp.corpo[1] {
						//parar de mandar mensagem desse tipo
						//avisar o controle quem é o líder atual
						controle <- actualLeader
					} else if temp.corpo[1] == temp.corpo[0] {
						temp.corpo[2] = actualLeader
						out <- temp
					} else {
						temp.corpo[1] = actualLeader
						out <- temp
					}

					fmt.Printf("%2d: O meu líder é o %d\n", TaskId, actualLeader)
				}
			case 5: //decide quem ganha a eleição e começa a avisar quem ganhou a eleição
				{
					fmt.Printf("CASE 5 - %2d: envia mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])

					if bFailed == false {
						ganhador := -1
						for i := 0; i < 3; i++ {
							if temp.corpo[i] > ganhador {
								ganhador = temp.corpo[i]
							}
						}

						actualLeader = ganhador

						temp.corpo[0] = actualLeader
						temp.tipo = 4

						fmt.Printf("CASE 5 - %2d: o processo ganhador foi o %d\n", TaskId, actualLeader)
						fmt.Printf("%2d: O meu líder é o %d\n", TaskId, actualLeader)
					}

					out <- temp
				}
			case 6: //fazer processo voltar
				{
					bFailed = false
					actualLeader = temp.corpo[0]
					controle <- 6
				}
			case 7: //matar os processos
				{
					fmt.Printf("%2d: matando o processo\n", TaskId)
					finish = true
				}
			default:
				{
					fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
					fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
					finish = true
				}
			}
		}

		if finish == true {
			break
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
