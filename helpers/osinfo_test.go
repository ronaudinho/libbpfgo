package helpers

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOSReleaseID_String(t *testing.T) {
	testCases := []struct {
		testName string
		in       uint32
		want     string
	}{
		{
			testName: "happy path",
			in:       1,
			want:     "ubuntu",
		},
		{
			testName: "no exists",
			in:       10,
			want:     "",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			id := OSReleaseID(tt.in)
			assert.Equal(t, tt.want, id.String())
		})
	}
}

func TestGetOSInfo(t *testing.T) {
	testCases := []struct {
		testName                  string
		osReleaseFilePath         string
		expectedOSReleaseFilePath string
		expectedOSReleaseID       OSReleaseID
		expectedError             error
	}{
		{
			testName:                  "env os-release filepath",
			osReleaseFilePath:         "testdata/os-release",
			expectedOSReleaseFilePath: "testdata/os-release",
			expectedOSReleaseID:       DEBIAN,
			expectedError:             nil,
		},
		{
			testName:                  "default os-release filepath",
			osReleaseFilePath:         "",
			expectedOSReleaseFilePath: "/etc/os-release",
			expectedError:             nil,
		},
		{
			testName:                  "non-existent os-release filepath",
			osReleaseFilePath:         "testdata/release",
			expectedOSReleaseFilePath: "testdata/release",
			expectedError:             errors.New("could not open LIBBPFGO_OSRELEASE_FILE testdata/release"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			var err error
			err = os.Setenv("LIBBPFGO_OSRELEASE_FILE", tt.osReleaseFilePath)
			assert.NoError(t, err)

			osInfo, err := GetOSInfo()
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				return
			}
			// covering unexpected uname error
			if err != nil {
				assert.ErrorContains(t, err, "uname")
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedOSReleaseFilePath, osInfo.GetOSReleaseFilePath())
			if tt.expectedOSReleaseID > 0 {
				assert.Equal(t, tt.expectedOSReleaseID, osInfo.GetOSReleaseID())
				assert.Equal(t, tt.expectedOSReleaseID.String(), osInfo.GetOSReleaseFieldValue(OS_ID))
			}
		})
	}
}

func TestOSInfo_CompareOSBaseKernelRelease(t *testing.T) {
	testCases := []struct {
		testName           string
		base               string
		given              string
		expectedComparison KernelVersionComparison
		expectedError      error
	}{
		{
			testName:           "older than",
			base:               "5.1.1",
			given:              "4.3.2",
			expectedComparison: KernelVersionOlder,
			expectedError:      nil,
		},
		{
			testName:           "equal",
			base:               "5.0",
			given:              "5.0",
			expectedComparison: KernelVersionEqual,
			expectedError:      nil,
		},
		{
			testName:           "newer than",
			base:               "3.1.1",
			given:              "4.3.2",
			expectedComparison: KernelVersionNewer,
			expectedError:      nil,
		},
		{
			testName:           "newer than (missing patch)",
			base:               "3.1",
			given:              "4.3.2",
			expectedComparison: KernelVersionNewer,
			expectedError:      nil,
		},
		{
			testName:           "newer than (missing minor and match)",
			base:               "3",
			given:              "4.3.2",
			expectedComparison: KernelVersionNewer,
			expectedError:      nil,
		},
		{
			testName:           "invalid, too many",
			base:               "3.0.0.0",
			given:              "4.3.2",
			expectedComparison: KernelVersionInvalid,
			expectedError:      errors.New("invalid base kernel version format: 3.0.0.0"),
		},
		{
			testName:           "invalid, not a number",
			base:               "X.5.4",
			given:              "4.3.2",
			expectedComparison: KernelVersionInvalid,
			expectedError:      errors.New("invalid base kernel version value: X.5.4 issue with: X"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			osInfo := &OSInfo{
				osReleaseFieldValues: map[OSReleaseField]string{
					OS_KERNEL_RELEASE: tt.base,
				},
			}
			comp, err := osInfo.CompareOSBaseKernelRelease(tt.given)
			assert.Equal(t, tt.expectedComparison, comp)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
