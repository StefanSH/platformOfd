package platformOfd

import (
	"fmt"
	"testing"
	"time"
)

func TestPlatformOfd_GetReceipts(t *testing.T) {
	//date, err := time.Parse("02.01.2006 15:04:59", "02.12.2019 23:59:59")
	//assert.Error(t, err)
	pOfd := PlatformOfd("+7 (916) 4212861", "cboxOcaI")
	receipts, _ := pOfd.GetReceipts(time.Now())

	fmt.Print(receipts)
}
