package druid

type DruidSupervisorSepc struct {
	DataSchema struct {
		DataSource    string `json:"dataSource"`
		TimestampSpec struct {
			Column       string      `json:"column"`
			Format       string      `json:"format"`
			MissingValue interface{} `json:"missingValue"`
		} `json:"timestampSpec"`
		DimensionsSpec struct {
			Dimensions []struct {
				Type               string `json:"type"`
				Name               string `json:"name"`
				MultiValueHandling string `json:"multiValueHandling"`
				CreateBitmapIndex  bool   `json:"createBitmapIndex"`
			} `json:"dimensions"`
			DimensionExclusions  []string `json:"dimensionExclusions"`
			IncludeAllDimensions bool     `json:"includeAllDimensions"`
		} `json:"dimensionsSpec"`
		MetricsSpec []struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"metricsSpec"`
		GranularitySpec struct {
			Type               string        `json:"type"`
			SegmentGranularity string        `json:"segmentGranularity"`
			QueryGranularity   string        `json:"queryGranularity"`
			Rollup             bool          `json:"rollup"`
			Intervals          []interface{} `json:"intervals"`
		} `json:"granularitySpec"`
		TransformSpec struct {
			Filter     interface{} `json:"filter"`
			Transforms []struct {
				Type       string `json:"type"`
				Name       string `json:"name"`
				Expression string `json:"expression"`
			} `json:"transforms"`
		} `json:"transformSpec"`
	} `json:"dataSchema"`
	IoConfig struct {
		Topic       string `json:"topic"`
		InputFormat struct {
			Type             string `json:"type"`
			AvroBytesDecoder struct {
				Type     string      `json:"type"`
				URL      string      `json:"url"`
				Capacity int64       `json:"capacity"`
				Urls     interface{} `json:"urls"`
				Config   interface{} `json:"config"`
				Headers  interface{} `json:"headers"`
			} `json:"avroBytesDecoder"`
			BinaryAsString      bool `json:"binaryAsString"`
			ExtractUnionsByType bool `json:"extractUnionsByType"`
		} `json:"inputFormat"`
		Replicas           int    `json:"replicas"`
		TaskCount          int    `json:"taskCount"`
		TaskDuration       string `json:"taskDuration"`
		ConsumerProperties struct {
			BootstrapServers string `json:"bootstrap.servers"`
		} `json:"consumerProperties"`
		AutoScalerConfig                  interface{} `json:"autoScalerConfig"`
		PollTimeout                       int         `json:"pollTimeout"`
		StartDelay                        string      `json:"startDelay"`
		Period                            string      `json:"period"`
		UseEarliestOffset                 bool        `json:"useEarliestOffset"`
		CompletionTimeout                 string      `json:"completionTimeout"`
		LateMessageRejectionPeriod        interface{} `json:"lateMessageRejectionPeriod"`
		EarlyMessageRejectionPeriod       interface{} `json:"earlyMessageRejectionPeriod"`
		LateMessageRejectionStartDateTime interface{} `json:"lateMessageRejectionStartDateTime"`
		ConfigOverrides                   interface{} `json:"configOverrides"`
		IdleConfig                        interface{} `json:"idleConfig"`
		Stream                            string      `json:"stream"`
		UseEarliestSequenceNumber         bool        `json:"useEarliestSequenceNumber"`
	} `json:"ioConfig"`
	TuningConfig struct {
		Type                string `json:"type"`
		AppendableIndexSpec struct {
			Type                    string `json:"type"`
			PreserveExistingMetrics bool   `json:"preserveExistingMetrics"`
		} `json:"appendableIndexSpec"`
		MaxRowsInMemory                int         `json:"maxRowsInMemory"`
		MaxBytesInMemory               int64       `json:"maxBytesInMemory"`
		SkipBytesInMemoryOverheadCheck bool        `json:"skipBytesInMemoryOverheadCheck"`
		MaxRowsPerSegment              int         `json:"maxRowsPerSegment"`
		MaxTotalRows                   interface{} `json:"maxTotalRows"`
		IntermediatePersistPeriod      string      `json:"intermediatePersistPeriod"`
		MaxPendingPersists             int         `json:"maxPendingPersists"`
		IndexSpec                      struct {
			Bitmap struct {
				Type                       string `json:"type"`
				CompressRunOnSerialization bool   `json:"compressRunOnSerialization"`
			} `json:"bitmap"`
			DimensionCompression     string `json:"dimensionCompression"`
			StringDictionaryEncoding struct {
				Type string `json:"type"`
			} `json:"stringDictionaryEncoding"`
			MetricCompression string `json:"metricCompression"`
			LongEncoding      string `json:"longEncoding"`
		} `json:"indexSpec"`
		IndexSpecForIntermediatePersists struct {
			Bitmap struct {
				Type                       string `json:"type"`
				CompressRunOnSerialization bool   `json:"compressRunOnSerialization"`
			} `json:"bitmap"`
			DimensionCompression     string `json:"dimensionCompression"`
			StringDictionaryEncoding struct {
				Type string `json:"type"`
			} `json:"stringDictionaryEncoding"`
			MetricCompression string `json:"metricCompression"`
			LongEncoding      string `json:"longEncoding"`
		} `json:"indexSpecForIntermediatePersists"`
		ReportParseExceptions               bool        `json:"reportParseExceptions"`
		HandoffConditionTimeout             int         `json:"handoffConditionTimeout"`
		ResetOffsetAutomatically            bool        `json:"resetOffsetAutomatically"`
		SegmentWriteOutMediumFactory        interface{} `json:"segmentWriteOutMediumFactory"`
		WorkerThreads                       interface{} `json:"workerThreads"`
		ChatThreads                         interface{} `json:"chatThreads"`
		ChatRetries                         int         `json:"chatRetries"`
		HTTPTimeout                         string      `json:"httpTimeout"`
		ShutdownTimeout                     string      `json:"shutdownTimeout"`
		OffsetFetchPeriod                   string      `json:"offsetFetchPeriod"`
		IntermediateHandoffPeriod           string      `json:"intermediateHandoffPeriod"`
		LogParseExceptions                  bool        `json:"logParseExceptions"`
		MaxParseExceptions                  int64       `json:"maxParseExceptions"`
		MaxSavedParseExceptions             int         `json:"maxSavedParseExceptions"`
		SkipSequenceNumberAvailabilityCheck bool        `json:"skipSequenceNumberAvailabilityCheck"`
		RepartitionTransitionDuration       string      `json:"repartitionTransitionDuration"`
	} `json:"tuningConfig"`
}

type DruidSupervisorDataSchema struct {
	DataSource    string `json:"dataSource"`
	TimestampSpec struct {
		Column       string      `json:"column"`
		Format       string      `json:"format"`
		MissingValue interface{} `json:"missingValue"`
	} `json:"timestampSpec"`
	DimensionsSpec struct {
		Dimensions []struct {
			Type               string `json:"type"`
			Name               string `json:"name"`
			MultiValueHandling string `json:"multiValueHandling"`
			CreateBitmapIndex  bool   `json:"createBitmapIndex"`
		} `json:"dimensions"`
		DimensionExclusions  []string `json:"dimensionExclusions"`
		IncludeAllDimensions bool     `json:"includeAllDimensions"`
	} `json:"dimensionsSpec"`
	MetricsSpec []struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"metricsSpec"`
	GranularitySpec struct {
		Type               string        `json:"type"`
		SegmentGranularity string        `json:"segmentGranularity"`
		QueryGranularity   string        `json:"queryGranularity"`
		Rollup             bool          `json:"rollup"`
		Intervals          []interface{} `json:"intervals"`
	} `json:"granularitySpec"`
	TransformSpec struct {
		Filter     interface{} `json:"filter"`
		Transforms []struct {
			Type       string `json:"type"`
			Name       string `json:"name"`
			Expression string `json:"expression"`
		} `json:"transforms"`
	} `json:"transformSpec"`
}

type DruidSupervisorTuningConfig struct {
	Type                string `json:"type"`
	AppendableIndexSpec struct {
		Type                    string `json:"type"`
		PreserveExistingMetrics bool   `json:"preserveExistingMetrics"`
	} `json:"appendableIndexSpec"`
	MaxRowsInMemory                int         `json:"maxRowsInMemory"`
	MaxBytesInMemory               int64       `json:"maxBytesInMemory"`
	SkipBytesInMemoryOverheadCheck bool        `json:"skipBytesInMemoryOverheadCheck"`
	MaxRowsPerSegment              int         `json:"maxRowsPerSegment"`
	MaxTotalRows                   interface{} `json:"maxTotalRows"`
	IntermediatePersistPeriod      string      `json:"intermediatePersistPeriod"`
	MaxPendingPersists             int         `json:"maxPendingPersists"`
	IndexSpec                      struct {
		Bitmap struct {
			Type                       string `json:"type"`
			CompressRunOnSerialization bool   `json:"compressRunOnSerialization"`
		} `json:"bitmap"`
		DimensionCompression     string `json:"dimensionCompression"`
		StringDictionaryEncoding struct {
			Type string `json:"type"`
		} `json:"stringDictionaryEncoding"`
		MetricCompression string `json:"metricCompression"`
		LongEncoding      string `json:"longEncoding"`
	} `json:"indexSpec"`
	IndexSpecForIntermediatePersists struct {
		Bitmap struct {
			Type                       string `json:"type"`
			CompressRunOnSerialization bool   `json:"compressRunOnSerialization"`
		} `json:"bitmap"`
		DimensionCompression     string `json:"dimensionCompression"`
		StringDictionaryEncoding struct {
			Type string `json:"type"`
		} `json:"stringDictionaryEncoding"`
		MetricCompression string `json:"metricCompression"`
		LongEncoding      string `json:"longEncoding"`
	} `json:"indexSpecForIntermediatePersists"`
	ReportParseExceptions               bool        `json:"reportParseExceptions"`
	HandoffConditionTimeout             int         `json:"handoffConditionTimeout"`
	ResetOffsetAutomatically            bool        `json:"resetOffsetAutomatically"`
	SegmentWriteOutMediumFactory        interface{} `json:"segmentWriteOutMediumFactory"`
	WorkerThreads                       interface{} `json:"workerThreads"`
	ChatThreads                         interface{} `json:"chatThreads"`
	ChatRetries                         int         `json:"chatRetries"`
	HTTPTimeout                         string      `json:"httpTimeout"`
	ShutdownTimeout                     string      `json:"shutdownTimeout"`
	OffsetFetchPeriod                   string      `json:"offsetFetchPeriod"`
	IntermediateHandoffPeriod           string      `json:"intermediateHandoffPeriod"`
	LogParseExceptions                  bool        `json:"logParseExceptions"`
	MaxParseExceptions                  int64       `json:"maxParseExceptions"`
	MaxSavedParseExceptions             int         `json:"maxSavedParseExceptions"`
	SkipSequenceNumberAvailabilityCheck bool        `json:"skipSequenceNumberAvailabilityCheck"`
	RepartitionTransitionDuration       string      `json:"repartitionTransitionDuration"`
}

type DruidSupervisorIoConfig struct {
	Topic       string `json:"topic"`
	InputFormat struct {
		Type             string `json:"type"`
		AvroBytesDecoder struct {
			Type     string      `json:"type"`
			URL      string      `json:"url"`
			Capacity int64       `json:"capacity"`
			Urls     interface{} `json:"urls"`
			Config   interface{} `json:"config"`
			Headers  interface{} `json:"headers"`
		} `json:"avroBytesDecoder"`
		BinaryAsString      bool `json:"binaryAsString"`
		ExtractUnionsByType bool `json:"extractUnionsByType"`
	} `json:"inputFormat"`
	Replicas           int    `json:"replicas"`
	TaskCount          int    `json:"taskCount"`
	TaskDuration       string `json:"taskDuration"`
	ConsumerProperties struct {
		BootstrapServers string `json:"bootstrap.servers"`
	} `json:"consumerProperties"`
	AutoScalerConfig                  interface{} `json:"autoScalerConfig"`
	PollTimeout                       int         `json:"pollTimeout"`
	StartDelay                        string      `json:"startDelay"`
	Period                            string      `json:"period"`
	UseEarliestOffset                 bool        `json:"useEarliestOffset"`
	CompletionTimeout                 string      `json:"completionTimeout"`
	LateMessageRejectionPeriod        interface{} `json:"lateMessageRejectionPeriod"`
	EarlyMessageRejectionPeriod       interface{} `json:"earlyMessageRejectionPeriod"`
	LateMessageRejectionStartDateTime interface{} `json:"lateMessageRejectionStartDateTime"`
	ConfigOverrides                   interface{} `json:"configOverrides"`
	IdleConfig                        interface{} `json:"idleConfig"`
	Stream                            string      `json:"stream"`
	UseEarliestSequenceNumber         bool        `json:"useEarliestSequenceNumber"`
}

type DruidSupervisor struct {
	Type         string                      `json:"type"`
	DataSchema   DruidSupervisorDataSchema   `json:"dataSchema"`
	TuningConfig DruidSupervisorTuningConfig `json:"tuningConfig"`
	IoConfig     DruidSupervisorIoConfig     `json:"ioConfig"`
	Context      interface{}                 `json:"context"`
	Spec         DruidSupervisorSepc         `json:"spec"`
	Suspended    bool                        `json:"suspended"`
}
