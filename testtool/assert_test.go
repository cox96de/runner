package testtool

import (
	"os"
	"testing"

	"github.com/cox96de/runner/testtool/mock"
	"go.uber.org/mock/gomock"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeepEqualObject(t *testing.T) {
	t.Run("interface", func(t *testing.T) {
		DeepEqualObject(t, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}, "testdata/interface.json")
	})
	t.Run("non_interface", func(t *testing.T) {
		DeepEqualObject(t, corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}, "testdata/interface.json")
	})
	t.Run("not_exists", func(t *testing.T) {
		expectPath := "testdata/not_exists.json"
		_ = os.Remove(expectPath)
		controller := gomock.NewController(t)
		tb := mock.NewMockTestingT(controller)
		tb.EXPECT().Helper().AnyTimes()
		tb.EXPECT().Log(gomock.Any()).AnyTimes()
		tb.EXPECT().FailNow().AnyTimes()
		DeepEqualObject(tb, corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}, expectPath)
	})
}
