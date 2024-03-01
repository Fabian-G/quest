package hook_test

import (
	"testing"

	"github.com/Fabian-G/quest/hook"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/stretchr/testify/assert"
)

type trackerMock struct {
	activeTagsCalls int
	setTagsCalls    [][]string
	startCalls      [][]string
	stopCalls       int
	active          bool
	currentTags     []string
}

func (t *trackerMock) ActiveTags() ([]string, error) {
	t.activeTagsCalls++
	if !t.active {
		return nil, hook.ErrNoActiveTracking
	}
	return t.currentTags, nil
}

func (t *trackerMock) SetTags(tags []string) error {
	t.setTagsCalls = append(t.setTagsCalls, tags)
	if !t.active {
		return hook.ErrNoActiveTracking
	}
	t.currentTags = tags
	return nil
}

func (t *trackerMock) Start(tags []string) error {
	t.startCalls = append(t.startCalls, tags)
	t.active = true
	t.currentTags = tags
	return nil
}

func (t *trackerMock) Stop() error {
	t.stopCalls++
	t.active = false
	t.currentTags = nil
	return nil
}

func Test_TrackingClearsTrackingTagOnOthersAfterStarting(t *testing.T) {
	list := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDescription("Item 1 "+hook.TrackingTag+":latest")),
		todotxt.MustBuildItem(todotxt.WithDescription("Item 2 ")),
	)
	list.AddHook(hook.NewTracking(&trackerMock{}))

	err := list.Tasks()[1].SetTag(hook.TrackingTag, "latest")
	assert.NoError(t, err)

	_, ok := list.Tasks()[0].Tags()[hook.TrackingTag]
	assert.False(t, ok)
}

func Test_TrackingCallsStartWhenAddingTheTrackingTag(t *testing.T) {
	mock := &trackerMock{}
	list := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDescription("@aContext Item 1 +aProject an:ignored-tag")),
	)
	list.AddHook(hook.NewTracking(mock))

	err := list.Tasks()[0].SetTag(hook.TrackingTag, "latest")
	assert.NoError(t, err)

	assert.True(t, mock.active)
	assert.Equal(t, 1, len(mock.startCalls))
	// Also test for the correct order of tags here
	assert.Equal(t, []string{"+aProject", "@aContext", "Item 1"}, mock.startCalls[0])
}

func Test_TrackingTrimsContextAndProjectPrefixes(t *testing.T) {
	mock := &trackerMock{}
	list := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDescription("@aContext Item 1 +aProject an:ignored-tag")),
	)
	tracking := hook.NewTracking(mock)
	tracking.TrimContextPrefix = true
	tracking.TrimProjectPrefix = true
	list.AddHook(tracking)

	err := list.Tasks()[0].SetTag(hook.TrackingTag, "latest")
	assert.NoError(t, err)

	assert.True(t, mock.active)
	assert.Equal(t, 1, len(mock.startCalls))
	assert.Equal(t, []string{"aProject", "aContext", "Item 1"}, mock.startCalls[0])
}

func Test_RemovingTheTrackingTagStopsTheTracking(t *testing.T) {
	mock := &trackerMock{}
	list := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDescription("@aContext Item 1 +aProject an:ignored-tag")),
	)
	list.AddHook(hook.NewTracking(mock))

	err := list.Tasks()[0].SetTag(hook.TrackingTag, "latest")
	assert.NoError(t, err)
	err = list.Tasks()[0].SetTag(hook.TrackingTag, "")
	assert.NoError(t, err)

	assert.False(t, mock.active)
	assert.Equal(t, 1, mock.stopCalls)
}

func Test_ChangesWillPropagateToTheTracker(t *testing.T) {
	mock := &trackerMock{}
	list := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDescription("@aContext Item 1 +aProject an:ignored-tag")),
	)
	list.AddHook(hook.NewTracking(mock))

	err := list.Tasks()[0].SetTag(hook.TrackingTag, "latest")
	assert.NoError(t, err)
	err = list.Tasks()[0].EditDescription(list.Tasks()[0].Description() + " @anotherContext")
	assert.NoError(t, err)

	assert.True(t, mock.active)
	assert.Equal(t, 1, len(mock.startCalls))
	assert.Equal(t, 1, len(mock.setTagsCalls))
	assert.Equal(t, []string{"+aProject", "@aContext", "@anotherContext", "Item 1"}, mock.setTagsCalls[0])
}

func Test_OutOfBandChangesAreNotOverriden(t *testing.T) {
	mock := &trackerMock{}
	list := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDescription("@aContext Item 1 +aProject an:ignored-tag")),
	)
	list.AddHook(hook.NewTracking(mock))

	err := list.Tasks()[0].SetTag(hook.TrackingTag, "latest")
	assert.NoError(t, err)
	mock.Start([]string{"Some", "Unrelated", "Tags"})
	err = list.Tasks()[0].EditDescription(list.Tasks()[0].Description() + " @anotherContext")
	assert.NoError(t, err)

	assert.True(t, mock.active)
	assert.Equal(t, []string{"Some", "Unrelated", "Tags"}, mock.currentTags)
}

func Test_StartOverridesActiveTracking(t *testing.T) {
	mock := &trackerMock{}
	mock.Start([]string{"Some", "oob", "tracking"})
	list := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDescription("Item 1")),
	)
	list.AddHook(hook.NewTracking(mock))

	err := list.Tasks()[0].SetTag(hook.TrackingTag, "latest")
	assert.NoError(t, err)

	assert.True(t, mock.active)
	assert.Equal(t, 2, len(mock.startCalls))
	assert.Equal(t, []string{"Item 1"}, mock.startCalls[1])
}

func Test_RestartingTrackingCanBeDoneBySettingADifferentValue(t *testing.T) {
	mock := &trackerMock{}
	list := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDescription("Item 1 " + hook.TrackingTag + ":latest")),
	)
	list.AddHook(hook.NewTracking(mock))

	err := list.Tasks()[0].SetTag(hook.TrackingTag, "another-value")
	assert.NoError(t, err)

	assert.True(t, mock.active)
	assert.Equal(t, 1, len(mock.startCalls))
	assert.Equal(t, []string{"Item 1"}, mock.startCalls[0])
}

func Test_NotChangingTheTrackingTagDoesNotRestartTrackingOnChange(t *testing.T) {
	mock := &trackerMock{}
	list := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDescription("Item 1 " + hook.TrackingTag + ":latest")),
	)
	list.AddHook(hook.NewTracking(mock))

	err := list.Tasks()[0].SetTag("some-tag", "some-value")
	assert.NoError(t, err)

	assert.False(t, mock.active)
	assert.Equal(t, 0, len(mock.startCalls))
}

func Test_CompletingATaskStopsTracking(t *testing.T) {
	mock := &trackerMock{}
	list := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDescription("Item 1")),
	)
	list.AddHook(hook.NewTracking(mock))

	err := list.Tasks()[0].SetTag(hook.TrackingTag, "latest")
	assert.NoError(t, err)
	err = list.Tasks()[0].Complete()
	assert.NoError(t, err)

	assert.False(t, mock.active)
	assert.Equal(t, 1, len(mock.startCalls))
	assert.Equal(t, 1, mock.stopCalls)
}

func Test_RemovingATaskStopsTracking(t *testing.T) {
	mock := &trackerMock{}
	list := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDescription("Item 1")),
	)
	list.AddHook(hook.NewTracking(mock))

	err := list.Tasks()[0].SetTag(hook.TrackingTag, "latest")
	assert.NoError(t, err)
	err = list.Remove(1)
	assert.NoError(t, err)

	assert.False(t, mock.active)
	assert.Equal(t, 1, len(mock.startCalls))
	assert.Equal(t, 1, mock.stopCalls)
}
