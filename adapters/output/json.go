package output

import (
	"encoding/json"
	"fmt"
	"github.com/simplycubed/vulnscan/entities"
	"github.com/simplycubed/vulnscan/utils"
)

func JsonAdapter(command utils.Command, entity entities.Entity) error {
	out, err := json.MarshalIndent(entity, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(command.Output, string(out))
	return err
}
