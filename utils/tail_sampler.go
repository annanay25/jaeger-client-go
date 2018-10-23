// Copyright (c) 2018 The JaegerTracing authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"github.com/opentracing/opentracing-go"
	"math"
)

// Stores span statistics.
// Sum, mean, std dev.
type SpanStat struct {
	totalDuration                  int64
	num                            int64
	m_oldM, m_newM, m_oldS, m_newS float64
	// NOTE: Evaluate either a predefined or a historic standard deviation.
}

func (s SpanStat) Mean() float64 {
	s.Lock()
	if num > 0 {
		return m_newM
	} else {
		return 0.0
	}
	s.Unlock()
}

func (s SpanStat) Variance() float64 {
	s.Lock()
	if num > 1 {
		return m_newS / (num - 1)
	} else {
		return 0.0
	}
	s.Unlock()
}

func (s SpanStat) Push(x float64) (float64, float64) {
	s.Lock()
	num++

	if num == 1 {
		m_oldM = x
		m_newM = x
		m_oldS = 0.0
	} else {
		m_newM = m_oldM + (x-m_oldM)/num
		m_newS = m_oldS + (x-m_oldM)*(x-m_newM)

		// set up for next iteration
		m_oldM = m_newM
		m_oldS = m_newS
	}
	s.Unlock()
	return s.Mean(), s.Variance()
}

func (s SpanStat) StdDev() float64 {
	return Sqrt(s.Variance())
}

// Placeholder for all span statistics.
// Use an instance of TailSampler to find interesting spans.
type TailSampler struct {
	sync.Mutex

	// Identifier for a class of spans. Operation name is a good candidate.
	operation string
	// Store statistics of spans in a hashmap.
	spanRecorder map[string]SpanStat
}

// Returns a TailSampler instance.
func NewTailSampler(operation string) TailSampler {
	return &TailSampler{
		operation: operation,
	}
}

// Add a span (and its values) to the sampler.
func (t *TailSampler) addSpan(span Span) bool {
	// Update statistics for this category of spans.
	// Function returns -
	// bool - span is intereseting or not.
	// error - if there was an error with adding span to sampler.

	mean, variance := t.spanRecorder[span.operationName].Push(int64(span.duration / time.Millisecond))

	// Check if span.duration is non-interesting by checking limits.
	m := t.Mean()
	s := t.StdDev()

	toSend := true
	if m-2*s > span.duration && m+2*s < span.duration {
		// Its within limits, hence don't send this span.
		toSend = false
	}

	return toSend
}
