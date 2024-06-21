// Package metrics provides prometheus metrics
package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	// STATS_ARRAY_LEN size of array for statistical data
	STATS_ARRAY_LEN int = 100 //nolint
)

// metrics for the SayBackup request
var (
	vInfServerVersion *prometheus.GaugeVec

	// models
	modelCounter *prometheus.GaugeVec

	// api requests
	apiRequests *prometheus.GaugeVec

	// workloads
	launchedWorkloads *prometheus.GaugeVec

	// llama_prompt_streaming
	llmChannelCreated      *prometheus.GaugeVec
	llmChannelDeleted      *prometheus.GaugeVec
	llmChannelStreams      *prometheus.GaugeVec
	llmChannelSendTimeouts *prometheus.GaugeVec
	llmChannelReadTimeouts *prometheus.GaugeVec

	// websocket stuffs
	wsAuthFailures *prometheus.GaugeVec

	// healthz checks
	healthzSuccess *prometheus.GaugeVec
	healthzError   *prometheus.GaugeVec
)

var mut sync.Mutex

// NewMetrics initializes metrics
func NewMetrics() {
	vInfServerVersion = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "v_inf_server_version",
			Help: "the version of v-inf (metadata)",
		},
		[]string{
			"version",
			"commit",
			"date",
		},
	)

	modelCounter = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "v_inf_models_counter",
			Help: "counter for model usage",
		},
		[]string{
			"model",
		},
	)

	apiRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "v_inf_api_requests",
			Help: "requests counter for api",
		},
		[]string{
			"uri",
			"method",
		},
	)

	launchedWorkloads = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "v_inf_launched_workloads",
			Help: "number of launched workloads",
		},
		[]string{
			"type",
		},
	)

	llmChannelCreated = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "v_inf_llm_channels_created",
			Help: "number of created llm channels",
		},
		[]string{},
	)

	llmChannelDeleted = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "v_inf_llm_channels_deleted",
			Help: "number of deleted llm channels",
		},
		[]string{},
	)

	llmChannelStreams = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "v_inf_llm_channel_streams",
			Help: "number of streams on channels",
		},
		[]string{},
	)

	llmChannelSendTimeouts = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "v_inf_llm_channel_send_timeouts",
			Help: "counter for send on channel timeouts",
		},
		[]string{},
	)

	llmChannelReadTimeouts = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "v_inf_llm_channel_read_timeouts",
			Help: "counter for read on channel timeouts",
		},
		[]string{},
	)

	wsAuthFailures = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "v_inf_websocket_auth_failures",
			Help: "counter websocket auth failures",
		},
		[]string{},
	)

	healthzSuccess = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "v_inf_healthz_success",
			Help: "counter of healthz successes",
		},
		[]string{
			"check",
		},
	)

	healthzError = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "v_inf_healthz_error",
			Help: "counter of healthz errors",
		},
		[]string{
			"check",
		},
	)
}

// SetVCDNServerVersion sets label version
func SetVInfServerVersion(version, commit, date string) {
	vInfServerVersion.WithLabelValues(version, commit, date).Set(0)
}

// IncrementModels increments model usage counter
func IncrementModels(model string) {
	mut.Lock()
	defer mut.Unlock()

	modelCounter.WithLabelValues(model).Inc()
}

// IncrementAPIRequests increments api endpoint request
func IncrementAPIRequests(uri, method string) {
	mut.Lock()
	defer mut.Unlock()

	apiRequests.WithLabelValues(uri, method).Inc()
}

// IncrementWorkloads increments workload type
func IncrementWorkloads(wlType string) {
	mut.Lock()
	defer mut.Unlock()

	launchedWorkloads.WithLabelValues(wlType).Inc()
}

// IncrementLLMChannelsCreated increments created llm channels
func IncrementLLMChannelsCreated() {
	mut.Lock()
	defer mut.Unlock()

	llmChannelCreated.WithLabelValues().Inc()
}

// IncrementLLMChannelsDeleted increments deleted llm channels
func IncrementLLMChannelsDeleted() {
	mut.Lock()
	defer mut.Unlock()

	llmChannelDeleted.WithLabelValues().Inc()
}

// IncrementLLMChannelsStreams increments llm channel streams
func IncrementLLMChannelsStreams() {
	mut.Lock()
	defer mut.Unlock()

	llmChannelStreams.WithLabelValues().Inc()
}

// IncrementLLMChannelReadTimeouts increments read timeouts from channels
func IncrementLLMChannelReadTimeouts() {
	mut.Lock()
	defer mut.Unlock()

	llmChannelReadTimeouts.WithLabelValues().Inc()
}

// IncrementLLMChannelSendTimeouts increments send timeouts from channels
func IncrementLLMChannelSendTimeouts() {
	mut.Lock()
	defer mut.Unlock()

	llmChannelSendTimeouts.WithLabelValues().Inc()
}

// IncrementWSAuthFailures increments counter for websocket auth failures
func IncrementWSAuthFailures() {
	mut.Lock()
	defer mut.Unlock()

	wsAuthFailures.WithLabelValues().Inc()
}

// IncrementHealthzSuccess increments check for healthz success check
func IncrementHealthzSuccess(check string) {
	mut.Lock()
	defer mut.Unlock()

	healthzSuccess.WithLabelValues(check).Inc()
}

// IncrementHealthzError increments check for healthz error check
func IncrementHealthzError(check string) {
	mut.Lock()
	defer mut.Unlock()

	healthzError.WithLabelValues(check).Inc()
}
