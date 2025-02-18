import {Packs, Pack} from "@/interfaces/pack";
import axios from 'axios';
import {apiClient} from "@/services/api.service";
import {plainToInstance} from "class-transformer";


class PacksService {

  async fetchAllPacks(): Promise<Packs> {
    try {
      const response = await apiClient.get('/packs');
      return plainToInstance(Packs, response.data)
    } catch (error) {
      console.error('Error fetching packs:', error);
      throw error;
    }

  }

  async fetchOnePack(slug: string): Promise<Pack> {
    try {
      const response = await apiClient.get(`/packs/${slug}`);
      return plainToInstance(Pack, response.data)
    } catch (error) {
      console.error('Error fetching pack:', error);
      throw error;
    }
  }

}
