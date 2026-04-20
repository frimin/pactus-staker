package pipline

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/pactus-project/pactus/wallet"
)

func (p *piplineExecutor) ExportValidatorsCsv(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create csv file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	err = writer.Write([]string{"address", "label"})
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	labelIdx := 1
	for _, pip := range p.piplines {
		for _, wlt := range pip.walletList {
			for _, address := range wlt.ListAddresses(wallet.OnlyValidatorAddresses()) {
				err = writer.Write([]string{address.Address, strconv.Itoa(labelIdx)})
				if err != nil {
					return fmt.Errorf("failed to write row: %w", err)
				}
				labelIdx++
			}
		}
	}

	return nil
}
