package whatnot

/*
Whatnot depends on multiple subscribers to a notification channel

The following is an implementation of GO channel multiplexing
*/

func mergeRec(chans ...<-chan int) <-chan int {
	switch len(chans) {
	case 0:
		c := make(chan int)
		close(c)
		return c
	case 1:
		return chans[0]
	default:
		m := len(chans) / 2
		return mergeTwo(
			mergeRec(chans[:m]...),
			mergeRec(chans[m:]...))
	}
}

func mergeTwo(a, b <-chan int) <-chan int {
	c := make(chan int)

	go func() {
		defer close(c)
		for a != nil || b != nil {
			select {
			case v, ok := <-a:
				if !ok {
					a = nil
					//log.Printf("a is done")
					continue
				}
				c <- v
			case v, ok := <-b:
				if !ok {
					b = nil
					//log.Printf("b is done")
					continue
				}
				c <- v
			}
		}
	}()
	return c
}
