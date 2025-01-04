package multipipe

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MultiPipeTestSuite struct {
	suite.Suite
}

func TestMultiPipeTestSuite(t *testing.T) {
	suite.Run(t, new(MultiPipeTestSuite))
}

func (suite *MultiPipeTestSuite) TestRead() {
	mp := New()

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		pr, err := mp.PipeReader()
		suite.Nil(err)

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer pr.Close()

			b, err := io.ReadAll(pr)
			suite.Nil(err)
			suite.Equal(string(b), "KAMAHAMEHA!!")
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer mp.Close()

		n, err := io.Copy(mp, strings.NewReader("KAMAHAMEHA!!"))
		suite.Nil(err)
		suite.Equal(int64(12), n)
	}()

	wg.Wait()
}

func (suite *MultiPipeTestSuite) TestClose() {
	mp := New()

	pr, err := mp.PipeReader()
	suite.Nil(err)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer pr.Close()

		b, err := io.ReadAll(pr)
		suite.Nil(err)
		suite.Equal(string(b), "KAMAHAMEHA!!")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var n int64
		var err error

		n, err = io.Copy(mp, strings.NewReader("KAMAHAMEHA!!"))
		suite.Nil(err)
		suite.Equal(int64(12), n)

		mp.Close()

		n, err = io.Copy(mp, strings.NewReader("KAMAHAMEHA!!"))
		suite.Equal(ErrClosedMultiPipe, err)
		suite.Equal(int64(0), n)
	}()

	wg.Wait()
}

func (suite *MultiPipeTestSuite) TestCloseWithErrorUpstream() {
	expectedErr := fmt.Errorf("upstream error")

	mp := New()

	pr, err := mp.PipeReader()
	suite.Nil(err)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer pr.Close()

		b, err := io.ReadAll(pr)
		suite.Equal(expectedErr, err)
		suite.Equal(string(b), "KA")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		io.Copy(mp, strings.NewReader("KA"))

		mp.CloseWithError(expectedErr)
	}()

	wg.Wait()
}

func (suite *MultiPipeTestSuite) TestCloseWithErrorDownstream() {
	expectedErr := fmt.Errorf("downsteam error")

	mp := New()

	pr, err := mp.PipeReader()
	suite.Nil(err)

	pr.CloseWithError(expectedErr)

	n, err := io.Copy(mp, strings.NewReader("KA"))
	suite.Equal(expectedErr, err)
	suite.Equal(int64(0), n)

	suite.Nil(mp.CloseWithError(err))
}
