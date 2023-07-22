package testtool

import (
	"os"
	"testing"

	"github.com/cox96de/runner/testtool/mock"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAssertObject(t *testing.T) {
	t.Run("interface", func(t *testing.T) {
		AssertObject(t, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}, "testdata/interface.json", func(got, expect interface{}) {
			assert.DeepEqual(t, got, expect)
		})
	})
	t.Run("non_interface", func(t *testing.T) {
		AssertObject(t, corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}, "testdata/interface.json", func(got, expect interface{}) {
			assert.DeepEqual(t, got, expect)
		})
	})
	t.Run("not_exists", func(t *testing.T) {
		expectPath := "testdata/not_exists.json"
		_ = os.Remove(expectPath)
		t.Cleanup(func() {
			_ = os.Remove(expectPath)
		})
		controller := gomock.NewController(t)
		tb := mock.NewMockTestingT(controller)
		tb.EXPECT().Helper().AnyTimes()
		tb.EXPECT().Logf(gomock.Any(), gomock.Any()).AnyTimes()
		tb.EXPECT().FailNow().AnyTimes()
		AssertObject(tb, corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}, expectPath, func(got, expect interface{}) {
		})
	})
}

func TestDeepEqualObject(t *testing.T) {
	DeepEqualObject(t, "a", "testdata/deep_equal_object.json")
}
