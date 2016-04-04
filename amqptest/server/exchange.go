package server

import (
	"fmt"
	"cmd/go/testdata/testinternal3"
)

type Exchange interface {
	route(route string, d *Delivery) error
	addExchangeBinding(route string, e *Exchange)
	delExchangeBinding(route string)
	addBinding(route string, q *Queue)
	delBinding(route string)
}

type TopicExchange struct {
	name     string
	bindings map[string]*Queue
	exchangeBindings map[string]*Exchange
}

func NewTopicExchange(name string) *TopicExchange {
	return &TopicExchange{
		name:     name,
		bindings: make(map[string]*Queue),
		exchangeBindings: make(map[string]*Exchange),
	}
}

func (ex *TopicExchange) addExchangeBinding(route string, e *Exchange) {
	ex.exchangeBindings[route] = e
}

func (ex *TopicExchange) delExchangeBinding(route string) {
	delete(ex.exchangeBindings, route)
}

func (t *TopicExchange) addBinding(route string, q *Queue) {
	t.bindings[route] = q
}

func (t *TopicExchange) delBinding(route string) {
	delete(t.bindings, route)
}

func (t *TopicExchange) route(route string, d *Delivery) error {
	//This only delivers to the first matching item. Rabbitmq delivers to all matching items but only once.
	for bname, q := range t.bindings {
		if topicMatch(bname, route) {
			q.data <- d
			return nil
		}
	}
	for bname, ex := range t.exchangeBindings {
		if topicMatch(bname, route){
			if err := ex.(Exchange).route(route, d); err == nil{
				return nil
			}
		}
	}

	// The route doesn't match any binding, then will be discarded
	return fmt.Errorf("No bindings to route: %s", route)
}

type DirectExchange struct {
	name     string
	bindings map[string]*Queue
	exchangeBindings map[string]*Exchange
}

func NewDirectExchange(name string) *DirectExchange {
	return &DirectExchange{
		name:     name,
		bindings: make(map[string]*Queue),
		exchangeBindings: make(map[string]*Exchange),
	}
}

func (ex *DirectExchange) addExchangeBinding(route string, e *Exchange) {
	ex.exchangeBindings[route] = e
}

func (ex *DirectExchange) delExchangeBinding(route string) {
	delete(ex.exchangeBindings, route)
}

func (d *DirectExchange) addBinding(route string, q *Queue) {
	d.bindings[route] = q
}

func (d *DirectExchange) delBinding(route string) {
	delete(d.bindings, route)
}

func (d *DirectExchange) route(route string, delivery *Delivery) error {
	if q, ok := d.bindings[route]; ok {
		q.data <- delivery
		return nil
	}

	if ex, ok := d.exchangeBindings[route]; ok{
		if err := ex.(Exchange).route(route, d); err == nil{
			return nil
		}
	}

	return fmt.Errorf("No bindings to route: %s", route)

}
