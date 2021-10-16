/**
 * Copyright (C) The AetherFS Authors - All Rights Reserved
 * See LICENSE for more information.
 */

const protocolDefault = 'http'
const baseUrlDefault = 'localhost:8080'

export default class Client {
    static default() {
        return new Client()
    }

    constructor(opts) {
        let { protocol, baseUrl } = opts || {}

        protocol = protocol || protocolDefault
        baseUrl = baseUrl || process.env.AETHERFS_HOST
        baseUrl = baseUrl || baseUrlDefault

        this._baseUrl = `${protocol}://${baseUrl}`
    }

    ListDatasets() {
        return fetch(`${this._baseUrl}/api/v1/datasets`).then((resp) => resp.json())
    }

    ListTags(dataset) {
        return fetch(`${this._baseUrl}/api/v1/datasets/${dataset}/tags`).then((resp) => resp.json())
    }

    GetDataset(dataset, tag) {
        return fetch(`${this._baseUrl}/api/v1/datasets/${dataset}/tags/${tag}`).then((resp) => resp.json())
    }

    ReadFile(dataset, tag, file) {
        return fetch(this.FormatFileSystemURL(dataset, tag, file)).then((resp) => resp.text())
    }

    FormatFileSystemURL(dataset, tag, file) {
        return `${this._baseUrl}/fs/${dataset}/${tag}/${file}`
    }
}
