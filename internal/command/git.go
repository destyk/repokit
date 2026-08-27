package command

import (
	"context"
)

// GitClone performs a shallow clone of a specific ref.
func GitClone(
	ctx context.Context,
	url string,
	ref string,
	destination string,
) error {
	return Run(
		ctx,
		"",
		"git",
		[]string{
			"clone",
			"--quiet",
			"--depth",
			"1",
			"--single-branch",
			"--branch",
			ref,
			url,
			destination,
		},
		Quiet(),
	)
}

// GitFetchBranch fetches a specific branch.
func GitFetchBranch(
	ctx context.Context,
	repository string,
	branch string,
) error {
	return Run(
		ctx,
		repository,
		"git",
		[]string{
			"fetch",
			"--quiet",
			"--prune",
			"origin",
			branch,
		},
		Quiet(),
	)
}

// GitCheckoutBranch switches to a remote branch.
func GitCheckoutBranch(
	ctx context.Context,
	repository string,
	branch string,
) error {
	return Run(
		ctx,
		repository,
		"git",
		[]string{
			"checkout",
			"--quiet",
			"-B",
			branch,
			"origin/" + branch,
		},
		Quiet(),
	)
}

// GitFetchTags updates remote tags.
func GitFetchTags(
	ctx context.Context,
	repository string,
) error {
	return Run(
		ctx,
		repository,
		"git",
		[]string{
			"fetch",
			"--quiet",
			"--tags",
			"--prune",
			"origin",
		},
		Quiet(),
	)
}

// GitCheckoutTag checks out a specific tag in detached HEAD state.
func GitCheckoutTag(
	ctx context.Context,
	repository string,
	tag string,
) error {
	return Run(
		ctx,
		repository,
		"git",
		[]string{
			"checkout",
			"--quiet",
			"--detach",
			tag,
		},
		Quiet(),
	)
}

// GitPull updates the current branch using fast-forward only.
func GitPull(
	ctx context.Context,
	repository string,
) error {
	return Run(
		ctx,
		repository,
		"git",
		[]string{
			"pull",
			"--quiet",
			"--ff-only",
		},
		Quiet(),
	)
}
