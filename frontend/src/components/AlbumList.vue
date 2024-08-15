<template>
    <div class="album-list">
        <h1>Albums</h1>
        <div v-if="loading" class="loading">加载中...</div>
        <div v-else>
            <div v-for="album in albums" :key="album.id" class="album-item" @click="viewAlbum(album.id)">
                <h2>{{ album.name }}</h2>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import axiosInstance from '../services/axiosInstance';
import { useRouter } from 'vue-router';

const albums = ref([]);
const loading = ref(true);
const router = useRouter();

const fetchAlbums = async () => {
    try {
        const response = await axiosInstance.get('/api/v1/albums');
        albums.value = response.data.data;
    } catch (error) {
        console.error('Error fetching albums:', error);
    } finally {
        loading.value = false;
    }
};

const viewAlbum = (albumId) => {
    router.push({ name: 'AlbumDetail', params: { id: albumId } });
};

onMounted(() => {
    fetchAlbums();
});
</script>

<style scoped>
.album-list {
    padding: 10px;
}

.album-item {
    cursor: pointer;
    border: 1px solid #ccc;
    margin-bottom: 10px;
    padding: 10px;
    background-color: #f9f9f9;
}

.album-item:hover {
    background-color: #e9e9e9;
}

.loading {
    font-size: 1.2em;
    text-align: center;
    width: 100%;
}
</style>