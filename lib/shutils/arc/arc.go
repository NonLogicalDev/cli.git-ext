package arc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/NonLogicalDev/nld.cli.git-ext/lib/clitools"
	shutils "github.com/NonLogicalDev/nld.lib.go.shutils"

	"github.com/pkg/errors"
)

var PhabDiffRe = regexp.MustCompile(`(?m)^\s*Differential Revision:\s*(.+)`)

type jsA = []interface{}
type jsM = map[string]interface{}

func Cmd(args ...interface{}) *shutils.ShCMD {
	return shutils.Cmd("arc", args...)
}

func RevisionFromMessage(message string) string {
	groups := PhabDiffRe.FindStringSubmatch(message)
	if len(groups) == 0 {
		return "No Revision"
	}
	return groups[1]
}

func Diff(base, updateRevision string, extArgs []string) error {
	args := []interface{}{
		"diff", fmt.Sprintf("--base=%v", base),
	}

	if len(updateRevision) != 0 {
		args = append(args, fmt.Sprintf("--update=%v", updateRevision))
	}

	for _, a := range extArgs {
		args = append(args, a)
	}

	return shutils.
		Cmd("arc", args...).
		PipeStderr(os.Stderr).PipeStdout(os.Stdout).PipeStdin(os.Stdin).
		Run().Err()
}

func GetMSGForRevision(revisionID string) (string, error) {
	rx := regexp.MustCompile(`\d+`)

	rev_id_str := rx.FindString(revisionID)
	if len(rev_id_str) == 0 {
		return "", errors.Errorf("Incorrect revision name %v.", revisionID)
	}
	rev_id, err := strconv.Atoi(rev_id_str)
	clitools.UserError(err)

	request, _ := json.MarshalIndent(jsM{
		"revision_id": rev_id,
	}, "", "  ")

	res, err := ConduitCall("differential.getcommitmessage", request)
	if err != nil {
		return "", err
	}

	output := map[string]interface{}{}
	err = json.Unmarshal([]byte(res), &output)

	return output["response"].(string), nil
}

func ConduitCall(endpoint string, params []byte) (string, error) {
	request := bytes.NewReader(params)
	return shutils.
		Cmd("arc", "call-conduit", endpoint).
		PipeStdin(request).
		Run().Value()
}
