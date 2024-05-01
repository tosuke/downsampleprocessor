package downsampleprocessor

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

type downsampleProcessor struct {
	logger *zap.Logger

	duration       time.Duration
	maxCardinality int

	ctx       context.Context
	cancelCtx context.CancelCauseFunc
	shutdownC chan struct{}

	wg   sync.WaitGroup
	item chan pmetric.Metrics
	next consumer.Metrics
}

type worker struct {
	proc *downsampleProcessor
}

func newDownsampleProcessor(settings processor.CreateSettings, cfg *Config, next consumer.Metrics) *downsampleProcessor {
	return &downsampleProcessor{
		logger: settings.Logger,

		duration:       cfg.Duration,
		maxCardinality: cfg.MaxCardinality,

		cancelCtx: func(error) {},
		shutdownC: make(chan struct{}, 1),

		wg:   sync.WaitGroup{},
		item: make(chan pmetric.Metrics, 1),
		next: next,
	}
}

func newWorker(proc *downsampleProcessor) *worker {
	return &worker{proc: proc}
}

// Capabilities implements processor.Metrics.
func (dp *downsampleProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

// Start implements processor.Metrics.
func (dp *downsampleProcessor) Start(ctx context.Context, host component.Host) error {
	pctx := context.WithoutCancel(ctx)
	pctx, cancel := context.WithCancelCause(pctx)

	dp.cancelCtx = cancel

    w := newWorker(dp)
    w.start()

	return nil
}

// Shutdown implements processor.Metrics.
func (dp *downsampleProcessor) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		dp.cancelCtx(context.Cause(ctx))
	}()

	close(dp.shutdownC)

	dp.wg.Wait()
	return nil
}

// ConsumeMetrics implements processor.Metrics.
func (dp *downsampleProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	dp.item <- md
	return nil
}

var _ processor.Metrics = (*downsampleProcessor)(nil)

func (w *worker) start() {
    w.proc.wg.Add(1)
    go func() {
        defer w.proc.wg.Done()
        w.startLoop()
    }()
}

func (w *worker) startLoop() {
    w.proc.logger.Debug("worker started")
	for {
		select {
		case <-w.proc.shutdownC:
			return
		case md := <-w.proc.item:
			w.processMetrics(md)
		}
	}
}

func (w *worker) processMetrics(md pmetric.Metrics) {
	ctx := w.proc.ctx
	if err := w.proc.next.ConsumeMetrics(ctx, md); err != nil {
		w.proc.logger.Warn("failed to send", zap.Error(err))
	}
}
