package stub

import (
	"fmt"
	"os"
	"strings"
	"time"

	"testing"

	"github.com/openshift/cluster-samples-operator/pkg/apis/samplesresource/v1alpha1"
	"github.com/openshift/cluster-samples-operator/pkg/operatorstatus"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	restclient "k8s.io/client-go/rest"

	configv1 "github.com/openshift/api/config/v1"
	imagev1 "github.com/openshift/api/image/v1"
	operatorsv1api "github.com/openshift/api/operator/v1"
	templatev1 "github.com/openshift/api/template/v1"
)

var (
	conditions = []v1alpha1.SamplesResourceConditionType{v1alpha1.SamplesExist, v1alpha1.ImportCredentialsExist, v1alpha1.ConfigurationValid, v1alpha1.ImageChangesInProgress, v1alpha1.RemovedManagementStateOnHold, v1alpha1.MigrationInProgress, v1alpha1.ImportImageErrorsExist}
)

func TestWrongSampleResourceName(t *testing.T) {
	h, sr, event := setup()
	sr.Name = "foo"
	sr.Status.Conditions = nil
	err := h.Handle(event)
	validate(true, err, "", sr, []v1alpha1.SamplesResourceConditionType{}, []corev1.ConditionStatus{}, t)
}

func TestNoArchOrDist(t *testing.T) {
	h, sr, event := setup()
	err := h.Handle(event)
	statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr, conditions, statuses, t)
	err = h.Handle(event)
	statuses[0] = corev1.ConditionTrue
	validate(true, err, "", sr, conditions, statuses, t)
}

func TestWithDist(t *testing.T) {
	distlist := []v1alpha1.SamplesDistributionType{
		v1alpha1.CentosSamplesDistribution,
		v1alpha1.RHELSamplesDistribution,
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            v1alpha1.SamplesRegistryCredentials,
			ResourceVersion: "a",
		},
	}
	credEvent := v1alpha1.Event{Object: secret}

	for _, dist := range distlist {
		h, sr, event := setup()
		sr.Spec.InstallType = dist
		if dist == v1alpha1.RHELSamplesDistribution {
			fakesecretclient := h.secretclientwrapper.(*fakeSecretClientWrapper)
			fakesecretclient.s = secret
			err := h.Handle(event)
			statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
			validate(true, err, "", sr, conditions, statuses, t)
			err = h.Handle(credEvent)
			statuses = []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
			validate(true, err, "", sr, conditions, statuses, t)
			err = h.Handle(event)
			statuses = []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
			validate(true, err, "", sr, conditions, statuses, t)
			err = h.Handle(event)
			statuses = []corev1.ConditionStatus{corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
			validate(true, err, "", sr, conditions, statuses, t)
		} else {
			err := h.Handle(event)
			statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
			validate(true, err, "", sr, conditions, statuses, t)
			err = h.Handle(event)
			statuses = []corev1.ConditionStatus{corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
			validate(true, err, "", sr, conditions, statuses, t)
		}
	}
}

func TestWithArchDist(t *testing.T) {
	distlist := []v1alpha1.SamplesDistributionType{
		v1alpha1.CentosSamplesDistribution,
		v1alpha1.RHELSamplesDistribution,
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            v1alpha1.SamplesRegistryCredentials,
			ResourceVersion: "a",
		},
	}
	credEvent := v1alpha1.Event{Object: secret}

	for _, dist := range distlist {
		h, sr, event := setup()
		sr.Spec.Architectures = []string{
			v1alpha1.X86Architecture,
		}
		sr.Spec.InstallType = dist
		if dist == v1alpha1.RHELSamplesDistribution {
			fakesecretclient := h.secretclientwrapper.(*fakeSecretClientWrapper)
			fakesecretclient.s = secret
			err := h.Handle(event)
			statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
			validate(true, err, "", sr,
				conditions,
				statuses, t)
			err = h.Handle(credEvent)
			statuses = []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
			validate(true, err, "", sr,
				conditions,
				statuses, t)
			err = h.Handle(event)
			statuses = []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
			validate(true, err, "", sr,
				conditions,
				statuses, t)
			err = h.Handle(event)
			statuses = []corev1.ConditionStatus{corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
			validate(true, err, "", sr,
				conditions,
				statuses, t)
		} else {
			mimic(&h, dist, x86OKDContentRootDir)
			err := h.Handle(event)
			statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
			validate(true, err, "", sr,
				conditions,
				statuses, t)
			err = h.Handle(event)
			statuses = []corev1.ConditionStatus{corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
			validate(true, err, "", sr,
				conditions,
				statuses, t)
		}
	}

	h, sr, event := setup()
	fakesecretclient := h.secretclientwrapper.(*fakeSecretClientWrapper)
	fakesecretclient.s = secret
	mimic(&h, v1alpha1.RHELSamplesDistribution, ppc64OCPContentRootDir)
	sr.Spec.InstallType = v1alpha1.RHELSamplesDistribution
	sr.Spec.Architectures = []string{
		v1alpha1.PPCArchitecture,
	}
	err := h.Handle(credEvent)
	statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr,
		conditions,
		statuses, t)
	err = h.Handle(event)
	statuses = []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr,
		conditions,
		statuses, t)
	err = h.Handle(event)
	statuses = []corev1.ConditionStatus{corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr,
		conditions,
		statuses, t)

	// verify cannot change arch and distro
	sr.ResourceVersion = "2"
	sr.Spec.InstallType = v1alpha1.CentosSamplesDistribution
	sr.Spec.Architectures = []string{
		v1alpha1.X86Architecture,
	}
	mimic(&h, v1alpha1.CentosSamplesDistribution, x86OCPContentRootDir)
	h.Handle(event)
	invalidConfig(t, "cannot change installtype from", sr.Condition(v1alpha1.ConfigurationValid))
	sr.Spec.InstallType = v1alpha1.RHELSamplesDistribution
	h.Handle(event)
	invalidConfig(t, "cannot change architectures from", sr.Condition(v1alpha1.ConfigurationValid))
}

func TestWithBadDist(t *testing.T) {
	h, sr, event := setup()
	sr.Spec.InstallType = v1alpha1.SamplesDistributionType("foo")
	h.Handle(event)
	invalidConfig(t, "invalid install type", sr.Condition(v1alpha1.ConfigurationValid))
}

func TestWithBadDistPPCArch(t *testing.T) {
	h, sr, event := setup()
	sr.Spec.InstallType = v1alpha1.SamplesDistributionType("foo")
	sr.Spec.Architectures = []string{
		v1alpha1.PPCArchitecture,
	}
	h.Handle(event)
	invalidConfig(t, "invalid install type", sr.Condition(v1alpha1.ConfigurationValid))
}

func TestWithArch(t *testing.T) {
	h, sr, event := setup()
	sr.Spec.Architectures = []string{
		v1alpha1.X86Architecture,
	}
	err := h.Handle(event)
	validate(true, err, "", sr, conditions, []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}, t)
}

func TestWithBadArch(t *testing.T) {
	h, sr, event := setup()
	sr.Spec.Architectures = []string{
		"bad",
	}
	h.Handle(event)
	invalidConfig(t, "architecture bad unsupported", sr.Condition(v1alpha1.ConfigurationValid))
}

func TestConfigurationValidCondition(t *testing.T) {
	h, sr, event := setup()
	err := h.Handle(event)
	validate(true, err, "", sr, conditions, []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}, t)
	sr.Spec.InstallType = "rhel8"
	sr.ResourceVersion = "2"
	err = h.Handle(event)
	invalidConfig(t, "invalid install type", sr.Condition(v1alpha1.ConfigurationValid))
	sr.Spec.InstallType = "rhel"
	sr.ResourceVersion = "3"
	err = h.Handle(event)
	invalidConfig(t, "cannot change installtype from centos to rhel", sr.Condition(v1alpha1.ConfigurationValid))
	sr.Spec.InstallType = "centos"
	sr.ResourceVersion = "4"
	err = h.Handle(event)
	validate(true, err, "", sr, conditions, []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}, t)
	sr.ResourceVersion = "5"
	err = h.Handle(event)
	validate(true, err, "", sr, conditions, []corev1.ConditionStatus{corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}, t)
}

func TestManagementState(t *testing.T) {
	h, sr, event := setup()
	iskeys := getISKeys()
	tkeys := getTKeys()
	mimic(&h, v1alpha1.CentosSamplesDistribution, x86OKDContentRootDir)
	sr.Spec.ManagementState = operatorsv1api.Unmanaged

	err := h.Handle(event)
	statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr, conditions, statuses, t)

	fakeisclient := h.imageclientwrapper.(*fakeImageStreamClientWrapper)
	for _, key := range iskeys {
		if _, ok := fakeisclient.upsertkeys[key]; ok {
			t.Fatalf("upserted imagestream while unmanaged %#v", sr)
		}
	}
	faketclient := h.templateclientwrapper.(*fakeTemplateClientWrapper)
	for _, key := range tkeys {
		if _, ok := faketclient.upsertkeys[key]; ok {
			t.Fatalf("upserted template while unmanaged %#v", sr)
		}
	}

	sr.ResourceVersion = "2"
	sr.Spec.ManagementState = operatorsv1api.Managed
	err = h.Handle(event)
	statuses = []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr, conditions, statuses, t)

	err = h.Handle(event)
	// event after in progress set to true, sets exists to true
	statuses = []corev1.ConditionStatus{corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr, conditions, statuses, t)

	// event after exists is true that should trigger samples upsert
	err = h.Handle(event)
	// event after in progress set to true
	statuses = []corev1.ConditionStatus{corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr, conditions, statuses, t)

	for _, key := range iskeys {
		if _, ok := fakeisclient.upsertkeys[key]; !ok {
			t.Fatalf("did not upsert imagestream while managed %#v", sr)
		}
	}
	for _, key := range tkeys {
		if _, ok := faketclient.upsertkeys[key]; !ok {
			t.Fatalf("did not upsert template while managed %#v", sr)
		}
	}

	// mimic a remove attempt occuring while we are progressing and the remove on hold getting set to true
	progressing := sr.Condition(v1alpha1.ImageChangesInProgress)
	progressing.Status = corev1.ConditionTrue
	sr.ConditionUpdate(progressing)
	sr.ResourceVersion = "3"
	sr.Spec.ManagementState = operatorsv1api.Removed
	err = h.Handle(event)
	statuses[4] = corev1.ConditionTrue
	validate(true, err, "", sr, conditions, statuses, t)

	// mimic samples events completing such that in progress set to false
	// then analyze resulting samplesresource event ... remove on hold should
	// also be false
	progressing.Status = corev1.ConditionFalse
	progressing.Reason = ""
	sr.ConditionUpdate(progressing)
	sr.ResourceVersion = "4"
	statuses[3] = corev1.ConditionFalse
	statuses[4] = corev1.ConditionFalse
	err = h.Handle(event)
	validate(true, err, "", sr, conditions, statuses, t)

}

func TestSkipped(t *testing.T) {
	h, sr, event := setup()
	iskeys := getISKeys()
	tkeys := getTKeys()
	sr.Spec.SkippedImagestreams = iskeys
	sr.Spec.SkippedTemplates = tkeys

	mimic(&h, v1alpha1.CentosSamplesDistribution, x86OKDContentRootDir)

	err := h.Handle(event)
	validate(true, err, "", sr, conditions, []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}, t)

	fakeisclient := h.imageclientwrapper.(*fakeImageStreamClientWrapper)
	for _, key := range iskeys {
		if _, ok := h.skippedImagestreams[key]; !ok {
			t.Fatalf("bad skipped processing for %s: %#v", key, h)
		}
		if _, ok := fakeisclient.upsertkeys[key]; ok {
			t.Fatalf("skipped is reached client %s: %#v", key, h)
		}
	}
	faketclient := h.templateclientwrapper.(*fakeTemplateClientWrapper)
	for _, key := range tkeys {
		if _, ok := h.skippedTemplates[key]; !ok {
			t.Fatalf("bad skipped processing for %s: %#v", key, h)
		}
		if _, ok := faketclient.upsertkeys[key]; ok {
			t.Fatalf("skipped t reached client %s: %#v", key, h)
		}
	}

}

func TestProcessed(t *testing.T) {
	dists := []v1alpha1.SamplesDistributionType{
		v1alpha1.CentosSamplesDistribution,
		v1alpha1.RHELSamplesDistribution,
	}

	for _, dist := range dists {
		h, sr, event := setup()
		event.Object = sr
		sr.Spec.InstallType = dist
		sr.Spec.SamplesRegistry = "foo.io"
		iskeys := getISKeys()
		tkeys := getTKeys()

		if dist == v1alpha1.CentosSamplesDistribution {
			mimic(&h, dist, x86OKDContentRootDir)
		}
		if dist == v1alpha1.RHELSamplesDistribution {
			mimic(&h, dist, x86OCPContentRootDir)
		}

		err := h.Handle(event)
		validate(true, err, "", sr,
			conditions,
			[]corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}, t)

		err = h.Handle(event)
		validate(true, err, "", sr,
			conditions,
			[]corev1.ConditionStatus{corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}, t)
		err = h.Handle(event)
		validate(true, err, "", sr,
			conditions,
			[]corev1.ConditionStatus{corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}, t)

		fakeisclient := h.imageclientwrapper.(*fakeImageStreamClientWrapper)
		for _, key := range iskeys {
			if _, ok := h.skippedImagestreams[key]; ok {
				t.Fatalf("bad skipped processing for %s: %#v", key, h)
			}
			if _, ok := fakeisclient.upsertkeys[key]; !ok {
				t.Fatalf("is did not reach client %s: %#v", key, h)
			}
		}
		is, _ := fakeisclient.Get("", "foo", metav1.GetOptions{})
		if is == nil || !strings.HasPrefix(is.Spec.DockerImageRepository, sr.Spec.SamplesRegistry) {
			t.Fatalf("stream repo not updated %#v, %#v", is, h)
		}
		is, _ = fakeisclient.Get("", "bar", metav1.GetOptions{})
		if is == nil || !strings.HasPrefix(is.Spec.DockerImageRepository, sr.Spec.SamplesRegistry) {
			t.Fatalf("stream repo not updated %#v, %#v", is, h)
		}

		faketclient := h.templateclientwrapper.(*fakeTemplateClientWrapper)
		for _, key := range tkeys {
			if _, ok := h.skippedTemplates[key]; ok {
				t.Fatalf("bad skipped processing for %s: %#v", key, h)
			}
			if _, ok := faketclient.upsertkeys[key]; !ok {
				t.Fatalf("t did not reach client %s: %#v", key, h)
			}
		}

		// make sure registries are updated after already updating from the defaults
		sr.Spec.SamplesRegistry = "bar.io"
		sr.ResourceVersion = "2"
		sr.Spec.InstallType = dist
		// fake out that the samples completed updating
		progressing := sr.Condition(v1alpha1.ImageChangesInProgress)
		progressing.Status = corev1.ConditionFalse
		sr.ConditionUpdate(progressing)

		err = h.Handle(event)
		validate(true, err, "", sr,
			conditions,
			[]corev1.ConditionStatus{corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}, t)
		is, _ = fakeisclient.Get("", "foo", metav1.GetOptions{})
		if is == nil || !strings.HasPrefix(is.Spec.DockerImageRepository, sr.Spec.SamplesRegistry) {
			t.Fatalf("stream repo not updated %#v, %#v", is, h)
		}
		is, _ = fakeisclient.Get("", "bar", metav1.GetOptions{})
		if is == nil || !strings.HasPrefix(is.Spec.DockerImageRepository, sr.Spec.SamplesRegistry) {
			t.Fatalf("stream repo not updated %#v, %#v", is, h)
		}

	}

}

func TestImageStreamEvent(t *testing.T) {
	h, sr, event := setup()
	mimic(&h, v1alpha1.CentosSamplesDistribution, x86OKDContentRootDir)
	err := h.Handle(event)
	statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr, conditions, statuses, t)

	tagVersion := int64(1)
	is := &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
			Annotations: map[string]string{
				v1alpha1.SamplesVersionAnnotation: v1alpha1.GitVersionString(),
			},
		},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				{
					Name:       "foo",
					Generation: &tagVersion,
				},
			},
		},
		Status: imagev1.ImageStreamStatus{
			Tags: []imagev1.NamedTagEventList{
				{
					Tag: "foo",
					Items: []imagev1.TagEvent{
						{
							Generation: tagVersion,
						},
					},
				},
			},
		},
	}

	sr.Status.InstallType = v1alpha1.CentosSamplesDistribution
	sr.Status.Architectures = []string{v1alpha1.X86Architecture}
	sr.Status.SkippedImagestreams = []string{}
	sr.Status.SkippedTemplates = []string{}
	h.processImageStreamWatchEvent(is, false)
	validate(true, err, "", sr, conditions, statuses, t)

	// now make sure when a non standard change event is not ignored and that we update
	// the imagestream ... to mimic, remove annotation
	delete(is.Annotations, v1alpha1.SamplesVersionAnnotation)
	fakeisclient := h.imageclientwrapper.(*fakeImageStreamClientWrapper)
	delete(fakeisclient.upsertkeys, "foo")
	h.processImageStreamWatchEvent(is, false)
	if _, ok := fakeisclient.upsertkeys["foo"]; !ok {
		t.Fatalf("is did not reach client %s: %#v", "foo", h)
	}

	// mimic img change condition event that sets exists to true
	err = h.Handle(event)
	statuses[0] = corev1.ConditionTrue
	validate(true, err, "", sr, conditions, statuses, t)

	// with the update above, now process both of the imagestreams and see the in progress condition
	// go false
	is.Annotations[v1alpha1.SamplesVersionAnnotation] = v1alpha1.GitVersionString()
	h.processImageStreamWatchEvent(is, false)
	validate(true, err, "", sr, conditions, statuses, t)
	statuses[3] = corev1.ConditionFalse
	is = &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
			Annotations: map[string]string{
				v1alpha1.SamplesVersionAnnotation: v1alpha1.GitVersionString(),
			},
		},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				{
					Name:       "bar",
					Generation: &tagVersion,
				},
			},
		},
		Status: imagev1.ImageStreamStatus{
			Tags: []imagev1.NamedTagEventList{
				{
					Tag: "bar",
					Items: []imagev1.TagEvent{
						{
							Generation: tagVersion,
						},
					},
				},
			},
		},
	}
	h.processImageStreamWatchEvent(is, false)
	validate(true, err, "", sr, conditions, statuses, t)
}

func TestTemplateEvent(t *testing.T) {
	h, sr, event := setup()
	mimic(&h, v1alpha1.CentosSamplesDistribution, x86OKDContentRootDir)
	err := h.Handle(event)
	statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr, conditions, statuses, t)
	err = h.Handle(event)
	statuses[0] = corev1.ConditionTrue
	validate(true, err, "", sr, conditions, statuses, t)

	template := &templatev1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bo",
		},
	}

	faketclient := h.templateclientwrapper.(*fakeTemplateClientWrapper)
	delete(faketclient.upsertkeys, "bo")
	h.processTemplateWatchEvent(template, false)
	if _, ok := faketclient.upsertkeys["bo"]; !ok {
		t.Fatalf("template did not reach client %s: %#v", "bo", h)
	}

}

func TestCreateDeleteSecretBeforeCR(t *testing.T) {
	h, sr, event := setup()
	h.crdwrapper.(*fakeCRDWrapper).sr = nil
	event.Object = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            v1alpha1.SamplesRegistryCredentials,
			ResourceVersion: "a",
		},
	}

	err := h.Handle(event)
	validate(false, err, "Received secret samples-registry-credentials but do not have the SamplesResource yet", sr,
		conditions,
		[]corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}, t)
	event.Deleted = true
	err = h.Handle(event)
	validate(false, err, "Received secret samples-registry-credentials but do not have the SamplesResource yet", sr,
		conditions,
		[]corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}, t)

	event.Deleted = false
	event.Object = sr
	h.crdwrapper.(*fakeCRDWrapper).sr = sr
	err = h.Handle(event)
	validate(true, err, "", sr,
		conditions,
		[]corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}, t)
	err = h.Handle(event)
	validate(true, err, "", sr,
		conditions,
		[]corev1.ConditionStatus{corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}, t)

}

func TestCreateDeleteSecretAfterCR(t *testing.T) {
	h, sr, event := setup()

	err := h.Handle(event)
	statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr, conditions, statuses, t)

	event.Object = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            v1alpha1.SamplesRegistryCredentials,
			ResourceVersion: "a",
		},
	}
	err = h.Handle(event)
	statuses[1] = corev1.ConditionTrue
	validate(true, err, "", sr, conditions, statuses, t)

	event.Deleted = true
	err = h.Handle(event)
	statuses[1] = corev1.ConditionFalse
	validate(true, err, "", sr, conditions, statuses, t)

}

func setup() (Handler, *v1alpha1.SamplesResource, v1alpha1.Event) {
	h := NewTestHandler()
	sr, _ := h.CreateDefaultResourceIfNeeded(nil)
	h.crdwrapper.(*fakeCRDWrapper).sr = sr
	return h, sr, v1alpha1.Event{Object: sr}
}

func TestSameSecret(t *testing.T) {
	h, sr, event := setup()

	err := h.Handle(event)
	statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr, conditions, statuses, t)

	event.Object = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            v1alpha1.SamplesRegistryCredentials,
			ResourceVersion: "a",
		},
	}
	err = h.Handle(event)
	statuses[1] = corev1.ConditionTrue
	validate(true, err, "", sr, conditions, statuses, t)

	err = h.Handle(event)
	validate(true, err, "", sr, conditions, statuses, t)
}

func TestSecretAPIError(t *testing.T) {
	h, sr, event := setup()
	err := h.Handle(event)
	statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr, conditions, statuses, t)

	event.Object = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            v1alpha1.SamplesRegistryCredentials,
			ResourceVersion: "a",
		},
	}
	fakesecretclient := h.secretclientwrapper.(*fakeSecretClientWrapper)
	fakesecretclient.err = fmt.Errorf("problemchangingsecret")
	err = h.Handle(event)
	statuses[1] = corev1.ConditionUnknown
	validate(false, err, "problemchangingsecret", sr, conditions, statuses, t)
}

func TestImageGetError(t *testing.T) {
	errors := []error{
		fmt.Errorf("getstreamerror"),
		kerrors.NewNotFound(schema.GroupResource{}, "getstreamerror"),
	}
	for _, iserr := range errors {
		h, sr, event := setup()
		event.Object = sr

		mimic(&h, v1alpha1.CentosSamplesDistribution, x86OKDContentRootDir)

		fakeisclient := h.imageclientwrapper.(*fakeImageStreamClientWrapper)
		fakeisclient.geterrors = map[string]error{"foo": iserr}

		statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
		err := h.Handle(event)
		if !kerrors.IsNotFound(iserr) {
			statuses[3] = corev1.ConditionUnknown
			validate(false, err, "getstreamerror", sr, conditions, statuses, t)
		} else {
			validate(true, err, "", sr, conditions, statuses, t)
		}
	}

}

func TestImageStreamImportError(t *testing.T) {
	two := int64(2)
	one := int64(1)
	streams := []*imagev1.ImageStream{
		&imagev1.ImageStream{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
				Annotations: map[string]string{
					v1alpha1.SamplesVersionAnnotation: v1alpha1.GitVersionString(),
				},
			},
			Spec: imagev1.ImageStreamSpec{
				Tags: []imagev1.TagReference{
					imagev1.TagReference{
						Name:       "1",
						Generation: &two,
					},
					imagev1.TagReference{
						Name:       "2",
						Generation: &one,
					},
				},
			},
			Status: imagev1.ImageStreamStatus{
				Tags: []imagev1.NamedTagEventList{
					imagev1.NamedTagEventList{
						Tag: "1",
						Items: []imagev1.TagEvent{
							imagev1.TagEvent{
								Generation: two,
							},
						},
					},
					imagev1.NamedTagEventList{
						Tag: "2",
						Conditions: []imagev1.TagEventCondition{
							imagev1.TagEventCondition{
								Generation: two,
								Status:     corev1.ConditionFalse,
								Message:    "Internal error occurred: unknown: Not Found",
								Reason:     "InternalError",
								Type:       imagev1.ImportSuccess,
							},
						},
					},
				},
			},
		},
		&imagev1.ImageStream{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
				Annotations: map[string]string{
					v1alpha1.SamplesVersionAnnotation: v1alpha1.GitVersionString(),
				},
			},
			Spec: imagev1.ImageStreamSpec{
				Tags: []imagev1.TagReference{
					imagev1.TagReference{
						Name:       "1",
						Generation: &two,
					},
					imagev1.TagReference{
						Name:       "2",
						Generation: &one,
					},
				},
			},
			Status: imagev1.ImageStreamStatus{
				Tags: []imagev1.NamedTagEventList{
					imagev1.NamedTagEventList{
						Tag: "1",
						Conditions: []imagev1.TagEventCondition{
							imagev1.TagEventCondition{
								Generation: two,
								Status:     corev1.ConditionFalse,
								Message:    "Internal error occurred: unknown: Not Found",
								Reason:     "InternalError",
								Type:       imagev1.ImportSuccess,
							},
						},
					},
					imagev1.NamedTagEventList{
						Tag: "2",
						Conditions: []imagev1.TagEventCondition{
							imagev1.TagEventCondition{
								Generation: two,
								Status:     corev1.ConditionFalse,
								Message:    "Internal error occurred: unknown: Not Found",
								Reason:     "InternalError",
								Type:       imagev1.ImportSuccess,
							},
						},
					},
				},
			},
		},
	}
	for _, is := range streams {
		h, sr, _ := setup()
		mimic(&h, v1alpha1.CentosSamplesDistribution, x86OKDContentRootDir)
		progressing := sr.Condition(v1alpha1.ImageChangesInProgress)
		progressing.Status = corev1.ConditionTrue
		sr.ConditionUpdate(progressing)
		err := h.processImageStreamWatchEvent(is, false)
		if err != nil {
			t.Fatalf("processImageStreamWatchEvent error %#v for stream %#v", err, is)
		}
		if sr.ConditionFalse(v1alpha1.ImportImageErrorsExist) {
			t.Fatalf("processImageStreamWatchEvent did not set import error to true %#v for stream %#v", sr, is)
		}

		importErr := sr.Condition(v1alpha1.ImportImageErrorsExist)
		if len(importErr.Reason) == 0 || !strings.Contains(importErr.Reason, is.Name) {
			t.Fatalf("processImageStreamWatchEvent did not set import error reason field %#v for stream %#v", sr, is)
		}
	}
}

func TestImageStreamImportErrorRecovery(t *testing.T) {
	two := int64(2)
	one := int64(1)
	stream := &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
			Annotations: map[string]string{
				v1alpha1.SamplesVersionAnnotation: v1alpha1.GitVersionString(),
			},
		},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name:       "1",
					Generation: &two,
				},
				imagev1.TagReference{
					Name:       "2",
					Generation: &one,
				},
			},
		},
		Status: imagev1.ImageStreamStatus{
			Tags: []imagev1.NamedTagEventList{
				imagev1.NamedTagEventList{
					Tag: "1",
					Items: []imagev1.TagEvent{
						imagev1.TagEvent{
							Generation: two,
						},
					},
				},
				imagev1.NamedTagEventList{
					Tag: "2",
					Items: []imagev1.TagEvent{
						imagev1.TagEvent{
							Generation: two,
						},
					},
				},
			},
		},
	}
	h, sr, _ := setup()
	mimic(&h, v1alpha1.CentosSamplesDistribution, x86OKDContentRootDir)
	importError := sr.Condition(v1alpha1.ImportImageErrorsExist)
	importError.Status = corev1.ConditionTrue
	importError.Reason = "foo "
	sr.ConditionUpdate(importError)
	err := h.processImageStreamWatchEvent(stream, false)
	if err != nil {
		t.Fatalf("processImageStreamWatchEvent error %#v", err)
	}
	if sr.ConditionTrue(v1alpha1.ImportImageErrorsExist) {
		t.Fatalf("processImageStreamWatchEvent did not set import error to false %#v", sr)
	}
	importErr := sr.Condition(v1alpha1.ImportImageErrorsExist)
	if len(importErr.Reason) > 0 && strings.Contains(importErr.Reason, stream.Name) {
		t.Fatalf("processImageStreamWatchEvent did not set import error reason field %#v", sr)
	}
}

func TestTemplateGetEreror(t *testing.T) {
	errors := []error{
		fmt.Errorf("gettemplateerror"),
		kerrors.NewNotFound(schema.GroupResource{}, "gettemplateerror"),
	}
	for _, terr := range errors {
		h, sr, event := setup()
		event.Object = sr

		mimic(&h, v1alpha1.CentosSamplesDistribution, x86OKDContentRootDir)

		faketclient := h.templateclientwrapper.(*fakeTemplateClientWrapper)
		faketclient.geterrors = map[string]error{"bo": terr}

		statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
		err := h.Handle(event)
		if !kerrors.IsNotFound(terr) {
			statuses[3] = corev1.ConditionUnknown
			validate(false, err, "gettemplateerror", sr, conditions, statuses, t)
		} else {
			validate(true, err, "", sr, conditions, statuses, t)
		}
	}

}

func TestDeletedCR(t *testing.T) {
	h, sr, event := setup()
	event.Deleted = true
	err := h.Handle(event)
	statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr, conditions, statuses, t)
}

func TestSameCR(t *testing.T) {
	h, sr, event := setup()
	sr.ResourceVersion = "a"

	// first pass on the resource creates the samples
	statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	err := h.Handle(event)
	validate(true, err, "", sr, conditions, statuses, t)

	err = h.Handle(event)
	statuses[0] = corev1.ConditionTrue
	validate(true, err, "", sr, conditions, statuses, t)

	err = h.Handle(event)
	validate(true, err, "", sr, conditions, statuses, t)

}

func TestBadTopDirList(t *testing.T) {
	h, sr, event := setup()
	fakefinder := h.Filefinder.(*fakeResourceFileLister)
	fakefinder.errors = map[string]error{x86OKDContentRootDir: fmt.Errorf("badtopdir")}
	err := h.Handle(event)
	statuses := []corev1.ConditionStatus{corev1.ConditionUnknown, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(false, err, "badtopdir", sr, conditions, statuses, t)
}

func TestBadSubDirList(t *testing.T) {
	h, sr, event := setup()
	mimic(&h, v1alpha1.CentosSamplesDistribution, x86OKDContentRootDir)
	fakefinder := h.Filefinder.(*fakeResourceFileLister)
	fakefinder.errors = map[string]error{x86OKDContentRootDir + "/imagestreams": fmt.Errorf("badsubdir")}
	err := h.Handle(event)
	statuses := []corev1.ConditionStatus{corev1.ConditionUnknown, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(false, err, "badsubdir", sr, conditions, statuses, t)
}

func TestBadTopLevelStatus(t *testing.T) {
	h, sr, event := setup()
	fakestatus := h.crdwrapper.(*fakeCRDWrapper)
	fakestatus.updateerr = fmt.Errorf("badsdkupdate")
	err := h.Handle(event)
	// with deferring sdk updates to the very end, the local object will still have valid statuses on it, even though the error
	// error returned by h.Handle indicates etcd was not updated
	statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(false, err, "badsdkupdate", sr, conditions, statuses, t)
}

func TestUnsupportedDistroChange(t *testing.T) {
	h, sr, event := setup()
	err := h.Handle(event)
	statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr, conditions, statuses, t)

	sr.Spec.InstallType = v1alpha1.RHELSamplesDistribution
	sr.ResourceVersion = "2"
	err = h.Handle(event)
	invalidConfig(t, "cannot change installtype", sr.Condition(v1alpha1.ConfigurationValid))
}

func TestUnsupportedArchChange(t *testing.T) {
	h, sr, event := setup()
	sr.Spec.Architectures = []string{v1alpha1.PPCArchitecture}
	sr.Spec.InstallType = v1alpha1.RHELSamplesDistribution
	// fake out secret import as shortcut ... we test secret events elsewhere
	cred := sr.Condition(v1alpha1.ImportCredentialsExist)
	cred.Status = corev1.ConditionTrue
	sr.ConditionUpdate(cred)
	err := h.Handle(event)
	statuses := []corev1.ConditionStatus{corev1.ConditionFalse, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionFalse, corev1.ConditionFalse}
	validate(true, err, "", sr, conditions, statuses, t)

	sr.Spec.Architectures = []string{v1alpha1.X86Architecture}
	sr.ResourceVersion = "2"

	err = h.Handle(event)
	invalidConfig(t, "cannot change architecture", sr.Condition(v1alpha1.ConfigurationValid))
}

func invalidConfig(t *testing.T, msg string, cfgValid *v1alpha1.SamplesResourceCondition) {
	if cfgValid.Status != corev1.ConditionFalse {
		t.Fatalf("config valid condition not false: %v", cfgValid)
	}
	if !strings.Contains(cfgValid.Message, msg) {
		t.Fatalf("wrong config valid message: %s", cfgValid.Message)
	}
}

func getISKeys() []string {
	return []string{"foo", "bar"}
}

func getTKeys() []string {
	return []string{"bo", "go"}
}

func mimic(h *Handler, dist v1alpha1.SamplesDistributionType, topdir string) {
	registry1 := "docker.io"
	registry2 := "docker.io"
	if dist == v1alpha1.RHELSamplesDistribution {
		registry1 = "registry.access.redhat.com"
		registry2 = "registry.redhat.io"
	}
	fakefile := h.Filefinder.(*fakeResourceFileLister)
	fakefile.files = map[string][]fakeFileInfo{
		topdir: {
			{
				name: "imagestreams",
				dir:  true,
			},
			{
				name: "templates",
				dir:  true,
			},
		},
		topdir + "/" + "imagestreams": {
			{
				name: "foo",
				dir:  false,
			},
			{
				name: "bar",
				dir:  false,
			},
		},
		topdir + "/" + "templates": {
			{
				name: "bo",
				dir:  false,
			},
			{
				name: "go",
				dir:  false,
			},
		},
	}
	fakeisgetter := h.Fileimagegetter.(*fakeImageStreamFromFileGetter)
	foo := &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "foo",
			Labels: map[string]string{},
		},
		Spec: imagev1.ImageStreamSpec{
			DockerImageRepository: registry1,
			Tags: []imagev1.TagReference{
				{
					// no Name field set on purpose, cover more code paths
					From: &corev1.ObjectReference{
						Kind: "DockerImage",
					},
				},
			},
		},
	}
	bar := &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "bar",
			Labels: map[string]string{},
		},
		Spec: imagev1.ImageStreamSpec{
			DockerImageRepository: registry2,
			Tags: []imagev1.TagReference{
				{
					From: &corev1.ObjectReference{
						Name: registry2,
						Kind: "DockerImage",
					},
				},
			},
		},
	}
	fakeisgetter.streams = map[string]*imagev1.ImageStream{
		topdir + "/imagestreams/foo": foo,
		topdir + "/imagestreams/bar": bar,
		"foo": foo,
		"bar": bar,
	}
	faketempgetter := h.Filetemplategetter.(*fakeTemplateFromFileGetter)
	bo := &templatev1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "bo",
			Labels: map[string]string{},
		},
	}
	gogo := &templatev1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "go",
			Labels: map[string]string{},
		},
	}
	faketempgetter.templates = map[string]*templatev1.Template{
		topdir + "/templates/bo": bo,
		topdir + "/templates/go": gogo,
		"bo": bo,
		"go": gogo,
	}

}

func validate(succeed bool, err error, errstr string, sr *v1alpha1.SamplesResource, statuses []v1alpha1.SamplesResourceConditionType, conditions []corev1.ConditionStatus, t *testing.T) {
	if succeed && err != nil {
		t.Fatal(err)
	}
	if len(statuses) != len(conditions) {
		t.Fatalf("bad input statuses %#v conditions %#v", statuses, conditions)
	}
	if !succeed {
		if err == nil {
			t.Fatalf("error should have been reported")
		}
		if !strings.Contains(err.Error(), errstr) {
			t.Fatalf("unexpected error: %v, expected: %v", err, errstr)
		}
	}
	if sr != nil {
		if len(sr.Status.Conditions) != len(statuses) {
			t.Fatalf("condition arrays different lengths got %v\n expected %v", sr.Status.Conditions, statuses)
		}
		for i, c := range statuses {
			if sr.Status.Conditions[i].Type != c {
				t.Fatalf("statuses in wrong order or different types have %#v\n expected %#v", sr.Status.Conditions, statuses)
			}
			if sr.Status.Conditions[i].Status != conditions[i] {
				t.Fatalf("unexpected for succeed %v have status condition %#v expected condition %#v and status %#v", succeed, sr.Status.Conditions[i], c, conditions[i])
			}
		}
	}
}

func NewTestHandler() Handler {
	h := Handler{}

	h.initter = &fakeInClusterInitter{}

	h.crdwrapper = &fakeCRDWrapper{}
	cvowrapper := fakeCVOWrapper{}
	cvowrapper.state = &configv1.ClusterOperator{
		TypeMeta: metav1.TypeMeta{
			APIVersion: configv1.SchemeGroupVersion.String(),
			Kind:       "ClusterOperator",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "goo",
			Namespace: "gaa",
		},
		Status: configv1.ClusterOperatorStatus{
			Conditions: []configv1.ClusterOperatorStatusCondition{},
		},
	}

	h.cvowrapper = operator.NewClusterOperatorHandler(nil)
	h.cvowrapper.ClusterOperatorWrapper = &cvowrapper

	h.namespace = "foo"

	h.skippedImagestreams = make(map[string]bool)
	h.skippedTemplates = make(map[string]bool)

	h.Fileimagegetter = &fakeImageStreamFromFileGetter{streams: map[string]*imagev1.ImageStream{}}
	h.Filetemplategetter = &fakeTemplateFromFileGetter{}
	h.Filefinder = &fakeResourceFileLister{files: map[string][]fakeFileInfo{}}

	h.imageclientwrapper = &fakeImageStreamClientWrapper{
		upsertkeys:   map[string]bool{},
		streams:      map[string]*imagev1.ImageStream{},
		geterrors:    map[string]error{},
		listerrors:   map[string]error{},
		upserterrors: map[string]error{},
	}
	h.templateclientwrapper = &fakeTemplateClientWrapper{
		upsertkeys:   map[string]bool{},
		templates:    map[string]*templatev1.Template{},
		geterrors:    map[string]error{},
		listerrors:   map[string]error{},
		upserterrors: map[string]error{},
	}
	h.secretclientwrapper = &fakeSecretClientWrapper{}

	h.imagestreamFile = make(map[string]string)
	h.templateFile = make(map[string]string)

	return h
}

type fakeFileInfo struct {
	name     string
	size     int64
	mode     os.FileMode
	modeTime time.Time
	sys      interface{}
	dir      bool
}

func (f *fakeFileInfo) Name() string {
	return f.name
}
func (f *fakeFileInfo) Size() int64 {
	return f.size
}
func (f *fakeFileInfo) Mode() os.FileMode {
	return f.mode
}
func (f *fakeFileInfo) ModTime() time.Time {
	return f.modeTime
}
func (f *fakeFileInfo) Sys() interface{} {
	return f.sys
}
func (f *fakeFileInfo) IsDir() bool {
	return f.dir
}

type fakeImageStreamFromFileGetter struct {
	streams map[string]*imagev1.ImageStream
	err     error
}

func (f *fakeImageStreamFromFileGetter) Get(fullFilePath string) (*imagev1.ImageStream, error) {
	if f.err != nil {
		return nil, f.err
	}
	is, _ := f.streams[fullFilePath]
	return is, nil
}

type fakeTemplateFromFileGetter struct {
	templates map[string]*templatev1.Template
	err       error
}

func (f *fakeTemplateFromFileGetter) Get(fullFilePath string) (*templatev1.Template, error) {
	if f.err != nil {
		return nil, f.err
	}
	t, _ := f.templates[fullFilePath]
	return t, nil
}

type fakeResourceFileLister struct {
	files  map[string][]fakeFileInfo
	errors map[string]error
}

func (f *fakeResourceFileLister) List(dir string) ([]os.FileInfo, error) {
	err, _ := f.errors[dir]
	if err != nil {
		return nil, err
	}
	files, ok := f.files[dir]
	if !ok {
		return []os.FileInfo{}, nil
	}
	// jump through some hoops for method has pointer receiver compile errors
	fis := make([]os.FileInfo, 0, len(files))
	for _, fi := range files {
		tfi := fakeFileInfo{
			name:     fi.name,
			size:     fi.size,
			mode:     fi.mode,
			modeTime: fi.modeTime,
			sys:      fi.sys,
			dir:      fi.dir,
		}
		fis = append(fis, &tfi)
	}
	return fis, nil
}

type fakeImageStreamClientWrapper struct {
	streams      map[string]*imagev1.ImageStream
	upsertkeys   map[string]bool
	geterrors    map[string]error
	upserterrors map[string]error
	listerrors   map[string]error
}

func (f *fakeImageStreamClientWrapper) Get(namespace, name string, opts metav1.GetOptions) (*imagev1.ImageStream, error) {
	err, _ := f.geterrors[name]
	if err != nil {
		return nil, err
	}
	is, _ := f.streams[name]
	return is, nil
}

func (f *fakeImageStreamClientWrapper) List(namespace string, opts metav1.ListOptions) (*imagev1.ImageStreamList, error) {
	err, _ := f.geterrors[namespace]
	if err != nil {
		return nil, err
	}
	list := &imagev1.ImageStreamList{}
	for _, is := range f.streams {
		list.Items = append(list.Items, *is)
	}
	return list, nil
}

func (f *fakeImageStreamClientWrapper) Create(namespace string, stream *imagev1.ImageStream) (*imagev1.ImageStream, error) {
	if stream == nil {
		return nil, nil
	}
	f.upsertkeys[stream.Name] = true
	err, _ := f.upserterrors[stream.Name]
	if err != nil {
		return nil, err
	}
	f.streams[stream.Name] = stream
	return stream, nil
}

func (f *fakeImageStreamClientWrapper) Update(namespace string, stream *imagev1.ImageStream) (*imagev1.ImageStream, error) {
	return f.Create(namespace, stream)
}

func (f *fakeImageStreamClientWrapper) Delete(namespace, name string, opts *metav1.DeleteOptions) error {
	return nil
}

func (f *fakeImageStreamClientWrapper) Watch(namespace string) (watch.Interface, error) {
	return nil, nil
}

type fakeTemplateClientWrapper struct {
	templates    map[string]*templatev1.Template
	upsertkeys   map[string]bool
	geterrors    map[string]error
	upserterrors map[string]error
	listerrors   map[string]error
}

func (f *fakeTemplateClientWrapper) Get(namespace, name string, opts metav1.GetOptions) (*templatev1.Template, error) {
	err, _ := f.geterrors[name]
	if err != nil {
		return nil, err
	}
	is, _ := f.templates[name]
	return is, nil
}

func (f *fakeTemplateClientWrapper) List(namespace string, opts metav1.ListOptions) (*templatev1.TemplateList, error) {
	err, _ := f.geterrors[namespace]
	if err != nil {
		return nil, err
	}
	list := &templatev1.TemplateList{}
	for _, is := range f.templates {
		list.Items = append(list.Items, *is)
	}
	return list, nil
}

func (f *fakeTemplateClientWrapper) Create(namespace string, t *templatev1.Template) (*templatev1.Template, error) {
	if t == nil {
		return nil, nil
	}
	f.upsertkeys[t.Name] = true
	err, _ := f.upserterrors[t.Name]
	if err != nil {
		return nil, err
	}
	f.templates[t.Name] = t
	return t, nil
}

func (f *fakeTemplateClientWrapper) Update(namespace string, t *templatev1.Template) (*templatev1.Template, error) {
	return f.Create(namespace, t)
}

func (f *fakeTemplateClientWrapper) Delete(namespace, name string, opts *metav1.DeleteOptions) error {
	return nil
}

func (f *fakeTemplateClientWrapper) Watch(namespace string) (watch.Interface, error) {
	return nil, nil
}

type fakeSecretClientWrapper struct {
	s   *corev1.Secret
	err error
}

func (f *fakeSecretClientWrapper) Create(namespace string, s *corev1.Secret) (*corev1.Secret, error) {
	if f.err != nil {
		return nil, f.err
	}
	return s, nil
}

func (f *fakeSecretClientWrapper) Update(namespace string, s *corev1.Secret) (*corev1.Secret, error) {
	if f.err != nil {
		return nil, f.err
	}
	return s, nil
}

func (f *fakeSecretClientWrapper) Delete(namespace, name string, opts *metav1.DeleteOptions) error {
	return f.err
}

func (f *fakeSecretClientWrapper) Get(namespace, name string) (*corev1.Secret, error) {
	return f.s, f.err
}

type fakeInClusterInitter struct{}

func (f *fakeInClusterInitter) init(h *Handler, restconfig *restclient.Config) {}

type fakeCRDWrapper struct {
	updateerr error
	createerr error
	geterr    error
	sr        *v1alpha1.SamplesResource
}

func (f *fakeCRDWrapper) Update(opcfg *v1alpha1.SamplesResource) error { return f.updateerr }

func (f *fakeCRDWrapper) Create(opcfg *v1alpha1.SamplesResource) error { return f.createerr }

func (f *fakeCRDWrapper) Get(name string) (*v1alpha1.SamplesResource, error) {
	return f.sr, f.geterr
}

type fakeCVOWrapper struct {
	updateerr error
	createerr error
	geterr    error
	state     *configv1.ClusterOperator
}

func (f *fakeCVOWrapper) UpdateStatus(state *configv1.ClusterOperator) error { return f.updateerr }

func (f *fakeCVOWrapper) Create(state *configv1.ClusterOperator) error { return f.createerr }

func (f *fakeCVOWrapper) Get(name string) (*configv1.ClusterOperator, error) {
	return f.state, f.geterr
}
