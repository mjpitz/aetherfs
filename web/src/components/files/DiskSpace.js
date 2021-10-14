/**
 * Copyright (C) The AetherFS Authors - All Rights Reserved
 * See LICENSE for more information.
 */

export default class DiskSpace extends Number {
    static BYTE = new DiskSpace(1)
    static KIBIBYTE = new DiskSpace(1 << 10)
    static MEBIBYTE = new DiskSpace(1 << 20)
    static GIBIBYTE = new DiskSpace(1 << 30)

    constructor(i) {
        super(i)
    }

    toString() {
        let v = this.valueOf(), l = 'B'

        if (v > DiskSpace.GIBIBYTE) {
            v = Math.floor(v / DiskSpace.GIBIBYTE)
            l = "GiB"
        } else if (this > DiskSpace.MEBIBYTE) {
            v = Math.floor(v / DiskSpace.MEBIBYTE)
            l = "MiB"
        } else if (this > DiskSpace.KIBIBYTE) {
            v = Math.floor(v / DiskSpace.KIBIBYTE)
            l = "KiB"
        }

        return `${v} ${l}`
    }
}
