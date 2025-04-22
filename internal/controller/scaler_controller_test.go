/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Scaler Controller", func() {
	Context("When reconciling a resource", func() {

		It("should successfully reconcile the resource", func() {

			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
var _ = Describe("shouldSkipResource", func() {
	Context("when the resource has no annotations", func() {
		It("should not skip the resource", func() {
			objMeta := &meta.ObjectMeta{}
			gomega.Expect(shouldSkipResource(objMeta)).To(gomega.BeFalse())
			Expect(shouldSkipResource(objMeta)).To(BeFalse())
		})
	})

	Context("when the resource has the 'kubescale/exclude' annotation set to 'true'", func() {
		It("should skip the resource", func() {
			meta := &meta.ObjectMeta{
				Annotations: map[string]string{
					"kubescale/exclude": "true",
				},
			}
			Expect(shouldSkipResource(meta)).To(BeTrue())
		})
	})

	Context("when the resource has the 'kubescale/exclude' annotation set to 'false'", func() {
		It("should not skip the resource", func() {
			meta := &meta.ObjectMeta{
				Annotations: map[string]string{
					"kubescale/exclude": "false",
				},
			}
			Expect(shouldSkipResource(meta)).To(BeFalse())
		})
	})

	Context("when the resource has the 'kubescale/exclude-until' annotation with a future timestamp", func() {
		It("should skip the resource", func() {
			futureTime := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
			meta := &meta.ObjectMeta{
				Annotations: map[string]string{
					"kubescale/exclude-until": futureTime,
				},
			}
			Expect(shouldSkipResource(meta)).To(BeTrue())
		})
	})

	Context("when the resource has the 'kubescale/exclude-until' annotation with a past timestamp", func() {
		It("should not skip the resource", func() {
			pastTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
			meta := &meta.ObjectMeta{
				Annotations: map[string]string{
					"kubescale/exclude-until": pastTime,
				},
			}
			Expect(shouldSkipResource(meta)).To(BeFalse())
		})
	})

	Context("when the resource has an invalid 'kubescale/exclude-until' annotation", func() {
		It("should not skip the resource", func() {
			meta := &meta.ObjectMeta{
				Annotations: map[string]string{
					"kubescale/exclude-until": "invalid-timestamp",
				},
			}
			Expect(shouldSkipResource(meta)).To(BeFalse())
		})
	})
})
