/**
 * Copyright (C) The AetherFS Authors - All Rights Reserved
 * See LICENSE for more information.
 */

import {createRouter, createWebHashHistory} from 'vue-router'
import DatasetDetails from '@/components/views/DatasetDetails.vue'
import DatasetList from '@/components/views/DatasetList.vue'
import DatasetOverview from '@/components/views/DatasetOverview.vue'
import TagDetails from '@/components/views/TagDetails.vue'

const routes = [
    ['/', 'Datasets', DatasetList],
    ['/dataset/:dataset', 'DatasetDetails', DatasetDetails, [
        {
            path: '',
            name: 'DatasetOverview',
            component: DatasetOverview,
        },
        {
            path: 'tag/:version',
            name: 'TagDetails',
            component: TagDetails,
            children: [
                {
                    path: 'tree/:path*',
                    name: 'FileTreeView',
                    component: TagDetails,
                },
            ],
        }
    ]],
    ['/dataset/:scope/:dataset', 'ScopedDatasetDetails', DatasetDetails, [
        {
            path: '',
            name: 'ScopedDatasetOverview',
            component: DatasetOverview,
        },
        {
            path: 'tag/:version',
            name: 'ScopedTagDetails',
            component: TagDetails,
            children: [
                {
                    path: 'tree/:path*',
                    name: 'ScopedFileTreeView',
                    component: TagDetails,
                },
            ],
        }
    ]],
].map(([path, name, component, children]) => ({path, name, component, children}))

const router = createRouter({
    history: createWebHashHistory(process.env.BASE_URL),
    routes,
})

export default router
