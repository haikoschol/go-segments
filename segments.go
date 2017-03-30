package segments

import (
	"context"
	"github.com/newrelic/go-agent"
	"net/http"
)

type Segment interface {
	End()
	SetResponse(*http.Response)
}

type NopSegment struct{}

func (NopSegment) End()                       {}
func (NopSegment) SetResponse(*http.Response) {}

type NewrelicApplication struct {
	newrelic newrelic.Application
}

func NewNewrelicApplication(name string, key string) (*NewrelicApplication, error) {
	nr, err := newrelic.NewApplication(newrelic.NewConfig(name, key))
	if err != nil {
		return nil, err
	}
	return &NewrelicApplication{nr}, nil
}

func (app *NewrelicApplication) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		txn := app.newrelic.StartTransaction(r.URL.Path, w, r)
		defer txn.End()

		ctx := context.WithValue(r.Context(), "newrelicTransaction", txn)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

type NewrelicDatastoreSegment struct {
	s newrelic.DatastoreSegment
}

func (n NewrelicDatastoreSegment) End() {
	n.s.End()
}
func (n NewrelicDatastoreSegment) SetResponse(*http.Response) {}

func StartDatabaseSegment(r *http.Request, table string, operation string) Segment {
	if r == nil {
		return NopSegment{}
	}
	ctx := r.Context()
	newrelicTransaction, ok := ctx.Value("newrelicTransaction").(newrelic.Transaction)

	if !ok {
		return NopSegment{}
	}

	return NewrelicDatastoreSegment{newrelic.DatastoreSegment{
		StartTime:  newrelic.StartSegmentNow(newrelicTransaction),
		Product:    newrelic.DatastorePostgres,
		Collection: table,
		Operation:  operation,
	}}
}

func StartRedisSegment(r *http.Request, operation string, key string) Segment {
	if r == nil {
		return NopSegment{}
	}
	ctx := r.Context()
	newrelicTransaction, ok := ctx.Value("newrelicTransaction").(newrelic.Transaction)

	if !ok {
		return NopSegment{}
	}

	return NewrelicDatastoreSegment{newrelic.DatastoreSegment{
		StartTime:          newrelic.StartSegmentNow(newrelicTransaction),
		Product:            newrelic.DatastoreRedis,
		Operation:          operation,
		ParameterizedQuery: operation + " " + key,
		QueryParameters: map[string]interface{}{
			"key": key,
		},
	}}
}

type NewrelicExternalSegment struct {
	s newrelic.ExternalSegment
}

func (n NewrelicExternalSegment) End() {
	n.s.End()
}
func (n NewrelicExternalSegment) SetResponse(r *http.Response) {
	n.s.Response = r
}

func StartExternalSegment(r *http.Request, req *http.Request) Segment {
	if r == nil {
		return NopSegment{}
	}
	ctx := r.Context()
	newrelicTransaction, ok := ctx.Value("newrelicTransaction").(newrelic.Transaction)

	if !ok {
		return NopSegment{}
	}

	return NewrelicExternalSegment{newrelic.StartExternalSegment(newrelicTransaction, req)}
}

type NewrelicSegment struct {
	s newrelic.Segment
}

func (n NewrelicSegment) End() {
	n.s.End()
}
func (n NewrelicSegment) SetResponse(r *http.Response) {}

func StartSegment(r *http.Request, name string) Segment {
	if r == nil {
		return NopSegment{}
	}
	ctx := r.Context()
	newrelicTransaction, ok := ctx.Value("newrelicTransaction").(newrelic.Transaction)
	if !ok {
		return NopSegment{}
	}

	return NewrelicSegment{newrelic.StartSegment(newrelicTransaction, name)}

}

func AddAttribute(r *http.Request, key string, value string) {
	if r == nil {
		return
	}
	ctx := r.Context()
	newrelicTransaction, ok := ctx.Value("newrelicTransaction").(newrelic.Transaction)
	if !ok {
		return
	}

	newrelicTransaction.AddAttribute(key, value)
}
