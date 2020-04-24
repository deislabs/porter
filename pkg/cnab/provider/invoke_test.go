package cnabprovider

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/claim"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ClaimWriting(t *testing.T) {

	type op struct {
		bun    *bundle.Bundle
		action string
		claim  string
	}
	type test struct {
		name   string
		in     op
		status string
		want   bool
	}

	d := NewTestRuntime(t)

	eClaim, err := claim.New("exists")
	require.NoError(t, err)
	eClaim.Update(claim.ActionInstall, claim.StatusSuccess)

	err = d.claims.Save(*eClaim)
	require.NoError(t, err)

	bun := &bundle.Bundle{
		Actions: map[string]bundle.Action{
			"blah": bundle.Action{
				Stateless: true,
			},
			"other": bundle.Action{
				Stateless: false,
			},
		},
	}

	tests := []test{
		{
			name: "stateless action, no claim should result in temp claim not written",
			in: op{
				bun,
				"blah",
				"nonexistent",
			},
			status: claim.StatusFailure,
			want:   false,
		},
		{
			name: "stateless action, existing claim should result in non temp claim and should be written",
			in: op{
				bun,
				"blah",
				"exists",
			},
			status: claim.StatusFailure,
			want:   true,
		},
		{
			name: "stateful action, existing claim should result in non temp claim and should be written",
			in: op{
				bun,
				"other",
				"exists",
			},
			status: claim.StatusFailure,
			want:   true,
		},
	}

	for _, tc := range tests {
		in := tc.in
		c, temp, err := d.getClaim(in.bun, in.action, in.claim)
		require.NoError(t, err)
		c.Result.Action = in.action
		c.Result.Status = tc.status
		err = d.writeClaim(temp, c)
		assert.NoError(t, err)

		fc, err := d.claims.Read(in.claim)
		if tc.want {
			assert.NoErrorf(t, err, "expected claim for %s", tc.name)
			assert.Equalf(t, in.action, fc.Result.Action, "expected action=%s for %s", in.action, tc.name)
			assert.Equalf(t, tc.status, fc.Result.Status, "expected status=%s for %s", tc.status, tc.name)
		} else {
			assert.Error(t, err, "expected no claim for %s", tc.name)
		}
	}
}

func Test_ClaimLoading(t *testing.T) {
	type input struct {
		bun    *bundle.Bundle
		action string
		claim  string
	}

	type result struct {
		claim *claim.Claim
		temp  bool
		err   error
	}

	type test struct {
		name string
		in   input
		want result
	}

	bun := &bundle.Bundle{
		Actions: map[string]bundle.Action{
			"blah": bundle.Action{
				Stateless: true,
			},
			"other": bundle.Action{
				Stateless: false,
			},
		},
	}

	eClaim, err := claim.New("exists")
	require.NoError(t, err)
	eClaim.Update(claim.ActionInstall, claim.StatusSuccess)

	d := NewTestRuntime(t)

	err = d.claims.Save(*eClaim)
	require.NoError(t, err)

	tests := []test{
		{
			name: "stateless action, no claim should result in temp claim",
			in: input{
				bun,
				"blah",
				"nonexistent",
			},
			want: result{
				temp: true,
				err:  nil,
				claim: &claim.Claim{
					Installation: "nonexistent",
					Bundle:       bun,
				},
			},
		},
		{
			name: "stateless action, existing claim should result in non temp claim",
			in: input{
				bun,
				"blah",
				"exists",
			},
			want: result{
				claim: eClaim,
				temp:  false,
				err:   nil,
			},
		},
		{
			name: "stateful action, existing claim should result in non temp claim",
			in: input{
				bun,
				"other",
				"exists",
			},
			want: result{
				claim: eClaim,
				temp:  false,
				err:   nil,
			},
		},
		{
			name: "stateful action, non exist claim should result in error",
			in: input{
				bun,
				"other",
				"nonexist",
			},
			want: result{
				claim: eClaim,
				temp:  false,
				err:   errors.Wrap(claim.ErrClaimNotFound, "could not load bundle instance nonexist"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := tc.in
			want := tc.want
			_, temp, err := d.getClaim(in.bun, in.action, in.claim)
			assert.Equalf(t, want.temp, temp, "getClaim returned an unexpected temporary flag")
			if want.err == nil {
				assert.NoErrorf(t, err, "getClaim failed")
			} else {
				assert.EqualErrorf(t, err, want.err.Error(), "getClaim returned an unexpected error")
			}
		})
	}
}

func TestInvoke_NoClaimBubblesUpError(t *testing.T) {
	r := NewTestRuntime(t)

	args := ActionArguments{
		Claim: "mybuns",
	}
	err := r.Invoke("custom-action", args)
	require.EqualError(t, err, "could not load bundle instance mybuns: Claim does not exist")
}
