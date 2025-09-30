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
	chans[3] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")

	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// mudar o processo 1 - canal de entrada 0 - para iniciar eleição

	temp.tipo = 1
	chans[0] <- temp
	fmt.Printf("Controle: manda 1 começar eleição\n")
	lider = <-in // receber confirmação
	fmt.Printf("Controle: confirmação - novo lider %d\n", lider)

	//revive o 0
	temp.tipo = 3
	chans[3] <- temp
	fmt.Printf("Controle: revive processo 0\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // esperar confirmação

	//mata o 1 e o 3
	temp.tipo = 2
	chans[0] <- temp
	fmt.Printf("Controle: mata processo 1\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // esperar confirmação

	chans[2] <- temp
	fmt.Printf("Controle: mata processo 3\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // esperar confirmação

	//manda o 0 começar eleição
	temp.tipo = 1
	chans[3] <- temp
	fmt.Printf("Controle: manda 0 começar eleição\n")
	lider = <-in // receber confirmação
	fmt.Printf("Controle: confirmação - novo lider %d\n", lider)

	// matar os outros processos com mensagens de término
	temp.tipo = 999
	for i := 0; i < 4; i++ {
		chans[i] <- temp
	}

	fmt.Printf("\n   Lider escolhido : %d \n", lider)
	fmt.Printf("\n   Processo controlador concluído\n")

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
					fmt.Printf("%2d: Propago Eleicao\n", TaskId)
					temp.corpo[TaskId] = TaskId
				} else {
					fmt.Printf("%2d: Propago Eleicao, Estou Morto\n", TaskId)
					temp.corpo[TaskId] = -1
				}
				out <- temp

			}
		case 1: // começa a eleição
			{
				if !bFailed {
					fmt.Printf("%2d: Comeco Eleicao\n", TaskId)
					temp.tipo = 0
					temp.corpo[TaskId] = TaskId

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

						fmt.Printf("%2d: Meu Lider e %2d\n", TaskId, actualLeader)

						out <- recebi

						controle <- novoLider
					} else {
						// Se não for uma mensagem de eleição, repassar
						out <- recebi
						controle <- -5
					}
				} else {
					fmt.Printf("%2d: TO MORTO, NAO POSSO\n", TaskId)
					controle <- -5
				}
			}
		case 2: // mata o processo
			{
				bFailed = true
				fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
				fmt.Printf("%2d: lider atual %d \n", TaskId, actualLeader)
				controle <- -5
			}
		case 3: // revive o processo
			{
				bFailed = false
				fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
				controle <- -5
			}
		case 4: // confirma lider
			{
				if !bFailed {
					actualLeader = temp.attLider
					fmt.Printf("%2d: Meu Lider -> %2d\n", TaskId, actualLeader)
					// Só repassa se não for o processo que iniciou a eleição
					// Usa o campo corpo para identificar quem iniciou
					iniciador := -1
					for i := 0; i < 4; i++ {
						if temp.corpo[i] == i {
							iniciador = i
							break
						}
					}
					if TaskId != iniciador {
						out <- temp
					}
				} else {
					fmt.Printf("%2d: Recebi confirmacao de lider mas estou morto\n", TaskId)
					// Processos mortos repassam para não bloquear
					iniciador := -1
					for i := 0; i < 4; i++ {
						if temp.corpo[i] == i {
							iniciador = i
							break
						}
					}
					if TaskId != iniciador {
						out <- temp
					}
				}
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
			}
		}
	}
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

	fmt.Println("\n   Processo controlador criado")

	wg.Wait() // Wait for the goroutines to finish\
}
