<template>
    <div class="album-list">
        <h1>Albums</h1>
        <button @click="openCreateAlbumModal" class="create-album-button">创建相册</button>

        <div v-if="loading" class="loading">加载中...</div>
        <div v-if="errorMessage" class="error">{{ errorMessage }}</div>
        <div v-else>
            <div v-for="album in albums" :key="album.id" class="album-item" @click="viewAlbum(album.id)"> <!-- 在这里添加 @click -->
                <h2>{{ album.name }}</h2>
                <button @click.stop="openEditAlbumModal(album)" class="edit-album-button">修改相册名</button>
            </div>
        </div>

        <div v-if="showModal" class="modal">
            <div class="modal-content">
                <span class="close" @click="closeModal">&times;</span>
                <h2>创建新相册</h2>
                <input type="text" v-model="newAlbumName" placeholder="输入相册名称" />
                <button @click="createAlbum">确认创建</button>
            </div>
        </div>

        <div v-if="showEditModal" class="modal">
            <div class="modal-content">
                <span class="close" @click="closeEditModal">&times;</span>
                <h2>修改相册名</h2>
                <input type="text" v-model="editedAlbumName" placeholder="输入新的相册名称" />
                <button @click="updateAlbum">确认修改</button>
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
const errorMessage = ref(null);
const showModal = ref(false);
const newAlbumName = ref('');
const showEditModal = ref(false);
const editedAlbumName = ref('');
const currentAlbumId = ref(null);
const router = useRouter();

const fetchAlbums = async () => {
    try {
        const response = await axiosInstance.get('/api/v1/albums');
        albums.value = response.data.data;
    } catch (error) {
        console.error('Error fetching albums:', error);
        errorMessage.value = '加载相册失败，请稍后再试。';
    } finally {
        loading.value = false;
    }
};

const viewAlbum = (albumId) => {
    router.push({ name: 'AlbumDetail', params: { id: albumId } });
};

const openCreateAlbumModal = () => {
    showModal.value = true;
    newAlbumName.value = ''; // 重置输入框
};

const closeModal = () => {
    showModal.value = false;
};

const createAlbum = async () => {
    if (!newAlbumName.value) {
        alert('相册名称不能为空');
        return;
    }
    
    try {
        const response = await axiosInstance.post('/api/v1/albums', {
            album_name: newAlbumName.value,
        });

        // 刷新相册列表
        albums.value.push(response.data.data);
        closeModal();
    } catch (error) {
        console.error('Error creating album:', error);
        alert('创建相册失败，请稍后再试。');
    }
};

const openEditAlbumModal = (album) => {
    editedAlbumName.value = album.name; // 设置当前相册名
    currentAlbumId.value = album.id; // 保存当前相册 ID
    showEditModal.value = true;
};

const closeEditModal = () => {
    showEditModal.value = false;
};

const updateAlbum = async () => {
    if (!editedAlbumName.value) {
        alert('相册名称不能为空');
        return;
    }
    
    try {
        await axiosInstance.put(`/api/v1/albums/${currentAlbumId.value}`, {
            album_name: editedAlbumName.value,
        });

        // 更新本地相册列表
        const album = albums.value.find(a => a.id === currentAlbumId.value);
        if (album) {
            album.name = editedAlbumName.value;
        }
        closeEditModal();
    } catch (error) {
        console.error('Error updating album:', error);
        alert('修改相册名失败，请稍后再试。');
    }
};

onMounted(() => {
    fetchAlbums();
});
</script>

<style scoped>
.album-list {
    padding: 10px;
}

.create-album-button,
.edit-album-button {
    margin: 5px;
    padding: 8px;
    background-color: #007bff;
    color: white;
    border: none;
    cursor: pointer;
}

.create-album-button:hover,
.edit-album-button:hover {
    background-color: #0056b3;
}

.album-item {
    border: 1px solid #ccc;
    margin-bottom: 10px;
    padding: 10px;
    background-color: #f9f9f9;
}

.loading {
    font-size: 1.2em;
    text-align: center;
    width: 100%;
}

.error {
    color: red;
    text-align: center;
    font-size: 1.2em;
}

.modal {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
}

.modal-content {
    background-color: white;
    padding: 20px;
    border-radius: 5px;
    text-align: center;
}

.close {
    cursor: pointer;
    float: right;
    font-size: 20px;
}
</style>
