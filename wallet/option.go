package wallet

type Options struct {
	allowNegative        bool
	overflowIfExists     bool
	accumulationIfExists bool
}

func (opt *Options) ConflictFlag() (flag int, err error) {
	if opt.overflowIfExists && opt.accumulationIfExists {
		err = ErrConflict

		return
	}

	if opt.accumulationIfExists {
		flag = 1
	} else if opt.overflowIfExists {
		flag = 2
	}

	return
}

type Option func(o *Options)

func optionNew(option ...Option) *Options {
	opts := &Options{}
	for _, o := range option {
		o(opts)
	}

	return opts
}

func AllowNegativeOption() Option {
	return func(d *Options) {
		d.allowNegative = true
	}
}

func OverflowIfExistsOption() Option {
	return func(d *Options) {
		d.overflowIfExists = true
	}
}

func AccumulationIfExistsOption() Option {
	return func(d *Options) {
		d.accumulationIfExists = true
	}
}
