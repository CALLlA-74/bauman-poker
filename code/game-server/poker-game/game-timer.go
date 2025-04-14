package pokergame

import "time"

type gameTimer struct {
	t         *time.Timer
	isAlarmed bool
}

func newGameTimer() *gameTimer {
	return &gameTimer{
		t:         nil,
		isAlarmed: false,
	}
}

/*
запускает таймер, если он не был запущен ранее. Сработает через <durationSec> секунд
*/
func (g *gameTimer) start(durationSec time.Duration) {
	if g.t != nil {
		return
	}

	g.reset()
	g.t = time.NewTimer(durationSec) // time.Duration(durationSec) * time.Second
	go func() {
		<-g.t.C
		g.isAlarmed = true
		g.t = nil
	}()
}

/*
останавливает таймер. Если он был остановлен ранее, то возвращается false, иначе -- true
*/
func (g *gameTimer) stop() bool {
	if g.t == nil {
		return false
	}
	f := g.t.Stop()
	g.t = nil
	return f
}

/*
устанавливает флаг срабатывания таймера в false
*/
func (g *gameTimer) reset() {
	g.isAlarmed = false
}
