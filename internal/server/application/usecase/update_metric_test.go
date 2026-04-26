package usecase

import (
	"errors"
	"testing"

	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
	sharedlog "github.com/a-aleesshin/metrics/internal/shared/port/logger"
)

type metricRepositoryStub struct {
	gaugeByName   *metric.Gauge
	counterByName *metric.Counter

	getGaugeErr    error
	getCounterErr  error
	saveGaugeErr   error
	saveCounterErr error

	savedGauge   *metric.Gauge
	savedCounter *metric.Counter
}

type nopLogger struct{}

func (nopLogger) Info(string, ...sharedlog.Field) {}

func (nopLogger) Error(string, ...sharedlog.Field) {}

func (m *metricRepositoryStub) GetGaugeByName(name metric.Name) (*metric.Gauge, error) {
	if m.getGaugeErr != nil {
		return nil, m.getGaugeErr
	}
	return m.gaugeByName, nil
}

func (m *metricRepositoryStub) SaveGauge(gauge *metric.Gauge) error {
	if m.saveGaugeErr != nil {
		return m.saveGaugeErr
	}
	m.savedGauge = gauge
	return nil
}

func (m *metricRepositoryStub) GetCounterByName(name metric.Name) (*metric.Counter, error) {
	if m.getCounterErr != nil {
		return nil, m.getCounterErr
	}
	return m.counterByName, nil
}

func (m *metricRepositoryStub) SaveCounter(counter *metric.Counter) error {
	if m.saveCounterErr != nil {
		return m.saveCounterErr
	}
	m.savedCounter = counter
	return nil
}

type snapshotSaverStub struct {
	calls int
	err   error
}

func (s *snapshotSaverStub) Execute() error {
	s.calls++
	return s.err
}

func TestUpdateMetric_Execute(t *testing.T) {
	existingGauge, _ := metric.NewGauge("gauge-id", "Alloc", 10.5)
	existingCounter, _ := metric.NewCounter("counter-id", "PollCount", 3)

	tests := []struct {
		name              string
		command           UpdateMetricCommand
		repo              *metricRepositoryStub
		logger            sharedlog.Logger
		wantErr           error
		wantGaugeName     string
		wantGaugeValue    float64
		wantCounterName   string
		wantCounterDelta  int64
		expectGaugeSave   bool
		expectCounterSave bool
	}{
		{
			name: "create_new_gauge",
			command: UpdateMetricCommand{
				Type:  "gauge",
				Name:  "Alloc",
				Value: "123.45",
			},
			repo:            &metricRepositoryStub{},
			logger:          nopLogger{},
			wantGaugeName:   "Alloc",
			wantGaugeValue:  123.45,
			expectGaugeSave: true,
		},
		{
			name: "update_existing_gauge",
			command: UpdateMetricCommand{
				Type:  "gauge",
				Name:  "Alloc",
				Value: "200.5",
			},
			repo: &metricRepositoryStub{
				gaugeByName: existingGauge,
			},
			logger:          nopLogger{},
			wantGaugeName:   "Alloc",
			wantGaugeValue:  200.5,
			expectGaugeSave: true,
		},
		{
			name: "create_new_counter",
			command: UpdateMetricCommand{
				Type:  "counter",
				Name:  "PollCount",
				Value: "7",
			},
			repo:              &metricRepositoryStub{},
			logger:            nopLogger{},
			wantCounterName:   "PollCount",
			wantCounterDelta:  7,
			expectCounterSave: true,
		},
		{
			name: "update_existing_counter",
			command: UpdateMetricCommand{
				Type:  "counter",
				Name:  "PollCount",
				Value: "5",
			},
			repo: &metricRepositoryStub{
				counterByName: existingCounter,
			},
			logger:            nopLogger{},
			wantCounterName:   "PollCount",
			wantCounterDelta:  8,
			expectCounterSave: true,
		},
		{
			name: "empty_metric_name",
			command: UpdateMetricCommand{
				Type:  "gauge",
				Name:  "",
				Value: "1",
			},
			repo:    &metricRepositoryStub{},
			logger:  nopLogger{},
			wantErr: metric.ErrNameEmpty,
		},
		{
			name: "unsupported_metric_type",
			command: UpdateMetricCommand{
				Type:  "histogram",
				Name:  "Alloc",
				Value: "1",
			},
			repo:    &metricRepositoryStub{},
			logger:  nopLogger{},
			wantErr: metric.ErrUnsupportedMetricType,
		},
		{
			name: "invalid_gauge_value",
			command: UpdateMetricCommand{
				Type:  "gauge",
				Name:  "Alloc",
				Value: "abc",
			},
			repo:    &metricRepositoryStub{},
			logger:  nopLogger{},
			wantErr: metric.ErrInvalidMetricValue,
		},
		{
			name: "invalid_counter_value",
			command: UpdateMetricCommand{
				Type:  "counter",
				Name:  "PollCount",
				Value: "abc",
			},
			repo:    &metricRepositoryStub{},
			logger:  nopLogger{},
			wantErr: metric.ErrInvalidMetricValue,
		},
		{
			name: "get_gauge_error",
			command: UpdateMetricCommand{
				Type:  "gauge",
				Name:  "Alloc",
				Value: "10",
			},
			repo: &metricRepositoryStub{
				getGaugeErr: errors.New("get gauge failed"),
			},
			logger:  nopLogger{},
			wantErr: errors.New("get gauge failed"),
		},
		{
			name: "save_counter_error",
			command: UpdateMetricCommand{
				Type:  "counter",
				Name:  "PollCount",
				Value: "1",
			},
			repo: &metricRepositoryStub{
				saveCounterErr: errors.New("save counter failed"),
			},
			logger:  nopLogger{},
			wantErr: errors.New("save counter failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			uc := NewUpdateMetric(tt.repo, tt.logger, nil)

			// Act
			err := uc.Execute(tt.command)

			// Assert
			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectGaugeSave {
				if tt.repo.savedGauge == nil {
					t.Fatal("expected gauge to be saved")
				}
				if tt.repo.savedGauge.Name().String() != tt.wantGaugeName {
					t.Fatalf("expected gauge name %q, got %q", tt.wantGaugeName, tt.repo.savedGauge.Name().String())
				}
				if tt.repo.savedGauge.Value() != tt.wantGaugeValue {
					t.Fatalf("expected gauge value %v, got %v", tt.wantGaugeValue, tt.repo.savedGauge.Value())
				}
			}

			if tt.expectCounterSave {
				if tt.repo.savedCounter == nil {
					t.Fatal("expected counter to be saved")
				}
				if tt.repo.savedCounter.Name().String() != tt.wantCounterName {
					t.Fatalf("expected counter name %q, got %q", tt.wantCounterName, tt.repo.savedCounter.Name().String())
				}
				if tt.repo.savedCounter.Delta() != tt.wantCounterDelta {
					t.Fatalf("expected counter delta %d, got %d", tt.wantCounterDelta, tt.repo.savedCounter.Delta())
				}
			}
		})
	}
}

func TestUpdateMetric_Execute_WithSnapshotSaver(t *testing.T) {
	repo := &metricRepositoryStub{}
	saver := &snapshotSaverStub{}
	uc := NewUpdateMetric(repo, nopLogger{}, saver)

	err := uc.Execute(UpdateMetricCommand{
		Type: "gauge", Name: "Alloc", Value: "1.23",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if saver.calls != 1 {
		t.Fatalf("expected saver calls=1, got %d", saver.calls)
	}
}

func TestUpdateMetric_Execute_SnapshotSaverNotCalledOnRepoSaveError(t *testing.T) {
	repo := &metricRepositoryStub{
		saveGaugeErr: errors.New("save gauge failed"),
	}
	saver := &snapshotSaverStub{}
	uc := NewUpdateMetric(repo, nopLogger{}, saver)

	err := uc.Execute(UpdateMetricCommand{
		Type: "gauge", Name: "Alloc", Value: "1.23",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if saver.calls != 0 {
		t.Fatalf("expected saver calls=0, got %d", saver.calls)
	}
}

func TestUpdateMetric_Execute_SnapshotSaverErrorIsNonFatal(t *testing.T) {
	repo := &metricRepositoryStub{}
	saver := &snapshotSaverStub{
		err: errors.New("snapshot saver failed"),
	}
	uc := NewUpdateMetric(repo, nopLogger{}, saver)

	err := uc.Execute(UpdateMetricCommand{
		Type: "gauge", Name: "Alloc", Value: "1.23",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if saver.calls != 1 {
		t.Fatalf("expected saver calls=1, got %d", saver.calls)
	}
}
