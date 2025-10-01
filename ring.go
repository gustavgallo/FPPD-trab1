// Código exemplo para o trabaho de sistemas distribuidos (eleicao em anel)
// By Cesar De Rose - 2022

package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo     int    // 0 para matar
	corpo    [4]int // conteudo da mensagem para colocar os ids (usar um tamanho ocmpativel com o numero de processos no anel)
	attLider int
}

var (
	chans = []chan mensagem{ // vetor de canais para formar o anel de eleicao - chan[0], chan[1] and chan[2] ...
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

	var temp mensagem
	var lider int

	// comandos para o anel iniciam aqui

	// mudar o processo 0 - canal de entrada 3 - para falho (defini mensagem tipo 2 pra isto)

	temp.tipo = 2
	fmt.Printf("Controle: mudar o processo 3 para falho\n")
	fmt.Println("===============================================")
	chans[2] <- temp

	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação
	fmt.Println("===============================================")

	// mudar o processo 0 - canal de entrada 3 - para iniciar eleição

	temp.tipo = 1
	fmt.Printf("Controle: manda 0 começar eleição\n")
	fmt.Println("===============================================")

	chans[3] <- temp
	lider = <-in // receber confirmação
	fmt.Printf("Controle: confirmação - novo lider %d\n", lider)
	fmt.Println("===============================================")

	//revive o 3
	temp.tipo = 3
	chans[2] <- temp
	fmt.Printf("Controle: revive processo 3\n")
	fmt.Println("===============================================")

	fmt.Printf("Controle: confirmação %d\n", <-in) // esperar confirmação
	fmt.Println("===============================================")

	//mata o 1 e o 2
	temp.tipo = 2
	chans[0] <- temp
	fmt.Printf("Controle: mata processo 1\n")
	fmt.Println("===============================================")

	fmt.Printf("Controle: confirmação %d\n", <-in) // esperar confirmação
	fmt.Println("===============================================")

	chans[1] <- temp
	fmt.Printf("Controle: mata processo 2\n")
	fmt.Println("===============================================")

	fmt.Printf("Controle: confirmação %d\n", <-in) // esperar confirmação
	fmt.Println("===============================================")

	//manda o 3 começar eleição
	temp.tipo = 1
	chans[2] <- temp
	fmt.Printf("Controle: manda 3 começar eleição\n")
	fmt.Println("===============================================")

	lider = <-in // receber confirmação
	fmt.Printf("Controle: confirmação - novo lider %d\n", lider)
	fmt.Println("===============================================")

	//mata todos processos, menos o 0
	temp.tipo = 2
	chans[0] <- temp
	fmt.Printf("Controle: mata processo 1\n")
	fmt.Println("===============================================")

	fmt.Printf("Controle: confirmação %d\n", <-in) // esperar confirmação
	fmt.Println("===============================================")

	chans[1] <- temp
	fmt.Printf("Controle: mata processo 2\n")
	fmt.Println("===============================================")

	fmt.Printf("Controle: confirmação %d\n", <-in) // esperar confirmação
	fmt.Println("===============================================")

	chans[2] <- temp
	fmt.Printf("Controle: mata processo 3\n")
	fmt.Println("===============================================")

	fmt.Printf("Controle: confirmação %d\n", <-in) // esperar confirmação
	fmt.Println("===============================================")

	//manda o 0 começar eleição
	temp.tipo = 1
	chans[3] <- temp
	fmt.Printf("Controle: manda 0 começar eleição\n")
	fmt.Println("===============================================")

	lider = <-in // receber confirmação
	fmt.Printf("Controle: confirmação - novo lider %d\n", lider)
	fmt.Println("===============================================")

	fmt.Printf("\n   Lider escolhido : %d \n", lider)
	fmt.Printf("\nProcesso controlador concluído, matando sub-rotinas\n")
	fmt.Println("===============================================")

	// matar todos os  processos com mensagens de término
	temp.tipo = 999
	for i := 0; i < 4; i++ {
		chans[i] <- temp
	}

}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()

	// variaveis locais que indicam se este processo é o lider e se esta ativo

	var actualLeader int
	var bFailed bool = false // todos inciam sem falha

	actualLeader = leader // indicação do lider veio por parâmatro

	for {
		temp := <-in // ler mensagem
		fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3])

		switch temp.tipo {

		// 0 == eleicao
		// 1 == inicia eleicao
		// 2 == mata
		// 3 == revive
		// 4 == termina eleicao
		// 999 == finaliza

		case 0: // eleição
			{
				if !bFailed {
					temp.corpo[TaskId] = TaskId
					fmt.Printf("%2d: Propago Eleição\n", TaskId)
					fmt.Printf("%2d: CORPO ATUALIZADO -> [ %d, %d, %d, %d ]\n", TaskId, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3])
					fmt.Println("-----------------------------------------------")
				} else {
					temp.corpo[TaskId] = -1
					fmt.Printf("%2d: Propago Eleição, Estou Morto\n", TaskId)
					fmt.Printf("%2d: CORPO ATUALIZADO -> [ %d, %d, %d, %d ]\n", TaskId, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3])
					fmt.Println("-----------------------------------------------")
				}
				out <- temp

			}
		case 1: // começa a eleição
			{
				if !bFailed {
					temp.corpo[TaskId] = TaskId

					fmt.Printf("%2d: Começo Eleição\n", TaskId)
					fmt.Printf("%2d: CORPO ATUALIZADO -> [ %d, %d, %d, %d ]\n", TaskId, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3])
					fmt.Println("-----------------------------------------------")
					temp.tipo = 0

					out <- temp

					// Esperar a mensagem voltar com todos os IDs
					recebi := <-in

					// Só processar se for uma mensagem de eleição que voltou
					if recebi.tipo == 0 {
						var novoLider int
						novoLider = recebi.corpo[0]
						for i := 1; i <= 3; i++ {
							if recebi.corpo[i] > novoLider && recebi.corpo[i] != -1 {
								novoLider = recebi.corpo[i]
							}
						}

						// Enviar mensagem de confirmação de líder
						recebi.tipo = 4
						recebi.attLider = novoLider
						actualLeader = novoLider

						fmt.Printf("%2d: Meu Lider é%2d\n", TaskId, actualLeader)
						fmt.Printf("%2d: Propago a escolha do Líder!\n", TaskId)
						fmt.Println("-----------------------------------------------")

						out <- recebi
						<-in // esperar o OK
						// informar o controle quem é o novo líder
						controle <- novoLider
					} else {
						// Se não for uma mensagem de eleição, repassar
						out <- recebi
						controle <- -5
					}
				} else {
					fmt.Printf("%2d: Estou morto, não posso realizar eleição\n", TaskId)
					controle <- -5
				}
			}
		case 2: // mata o processo
			{
				if !bFailed {
					bFailed = true
					fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
					fmt.Printf("%2d: lider atual %d \n", TaskId, actualLeader)
					fmt.Println("-----------------------------------------------")
					controle <- -5
				} else {
					fmt.Printf("%2d: já estou falho!! \n", TaskId)
					fmt.Printf("%2d: lider atual %d \n", TaskId, actualLeader)
					fmt.Println("-----------------------------------------------")
					controle <- -5

				}
			}
		case 3: // revive o processo
			{
				bFailed = false
				fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
				fmt.Println("-----------------------------------------------")
				controle <- -5
			}
		case 4: // confirma lider
			{
				fmt.Printf("%2d: Meu Lider -> %2d\n", TaskId, temp.attLider)
				fmt.Println("-----------------------------------------------")
				out <- temp
			}
		case 999: // terminar processo
			{
				fmt.Printf("%2d: terminei \n", TaskId)

				return
			}
		default:
			{
				fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
				fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
				fmt.Println("-----------------------------------------------")
			}
		}
	}
}

func main() {

	wg.Add(5) // Add a count of four, one for each goroutine

	// criar os processo do anel de eleicao

	go ElectionStage(0, chans[3], chans[0], 3) // este é o lider
	go ElectionStage(1, chans[0], chans[1], 3) // não é lider, é o processo 0
	go ElectionStage(2, chans[1], chans[2], 3) // não é lider, é o processo 0
	go ElectionStage(3, chans[2], chans[3], 3) // não é lider, é o processo 0

	fmt.Println("\n   Anel de processos criado")
	fmt.Println()

	// criar o processo controlador

	go ElectionControler(controle)

	fmt.Println("\n   Processo controlador criado")

	wg.Wait() // Wait for the goroutines to finish\
}
