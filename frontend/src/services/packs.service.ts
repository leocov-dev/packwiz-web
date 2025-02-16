import type {Packs} from "@/interfaces/pack";
import axios from 'axios';


class PacksService {

  private readonly API_URL = 'http://localhost:8080/api/v1/packwiz/pack';

  async fetchAllPacks(): Promise<Packs> {
    try {
      const response = await axios.get<Packs>(this.API_URL); // Use axios to make a GET request
      return response.data; // Return the packs data
    } catch (error) {
      console.error('Error fetching packs:', error);
      throw error; // Re-throw the error for handling outside
    }

  }

}
