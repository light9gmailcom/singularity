// Copyright (c) 2018, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package sources

import (
	"fmt"
	"path/filepath"

	"github.com/singularityware/singularity/src/pkg/build/types"
	"github.com/singularityware/singularity/src/pkg/image"
	"github.com/singularityware/singularity/src/pkg/sylog"
	"github.com/singularityware/singularity/src/pkg/util/loop"
)

// LocalConveyor only needs to hold the conveyor to have the needed data to pack
type LocalConveyor struct {
	src string
	b   *types.Bundle
}

type localPacker interface {
	Pack() (*types.Bundle, error)
}

// LocalConveyorPacker only needs to hold the conveyor to have the needed data to pack
type LocalConveyorPacker struct {
	LocalConveyor
	localPacker
}

func getLocalPacker(src string, b *types.Bundle) (localPacker, error) {
	imageObject, err := image.Init(src, false)
	if err != nil {
		return nil, err
	}

	info := new(loop.Info64)

	switch imageObject.Type {
	case image.SIF:
		sylog.Debugf("Packing from SIF")

		return &SIFPacker{
			srcfile: src,
			b:       b,
		}, nil
	case image.SQUASHFS:
		sylog.Debugf("Packing from Squashfs")

		info.Offset = imageObject.Offset
		info.SizeLimit = imageObject.Size

		return &SquashfsPacker{
			srcfile: src,
			b:       b,
			info:    info,
		}, nil
	case image.EXT3:
		sylog.Debugf("Packing from Ext3")

		info.Offset = imageObject.Offset
		info.SizeLimit = imageObject.Size

		return &Ext3Packer{
			srcfile: src,
			b:       b,
			info:    info,
		}, nil
	case image.SANDBOX:
		sylog.Debugf("Packing from Sandbox")

		return &SandboxPacker{
			srcdir: src,
			b:      b,
		}, nil
	default:
		return nil, fmt.Errorf("invalid image format")
	}
}

// Get just stores the source
func (cp *LocalConveyorPacker) Get(recipe types.Definition) (err error) {
	cp.src = filepath.Clean(recipe.Header["from"])

	//create bundle to build into
	cp.b, err = types.NewBundle("sbuild-local")
	if err != nil {
		return
	}

	cp.localPacker, err = getLocalPacker(cp.src, cp.b)
	return err
}
