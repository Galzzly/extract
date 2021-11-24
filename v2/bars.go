package extract

import (
	"time"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

func AddNewBar(p *mpb.Progress, file string, start time.Time) (b *mpb.Bar) {
	b = p.Add(
		int64(1),
		mpb.NewBarFiller(
			mpb.SpinnerStyle([]string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●", "∙∙∙"}...).PositionLeft(),
		),
		// mpb.BarFillerClearOnComplete(),
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name(file+":", decor.WC{W: len(file) + 2, C: decor.DidentRight}),
			decor.OnComplete(decor.Name("Extracting", decor.WCSyncSpaceR), "Done!"),
		),
	)
	return
}
