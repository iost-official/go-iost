package asset

func (this *AssetDataStructure) max_settle_volume(supply ShareType) ShareType {
	if this.max_settle_volume == 0 {
		return 0
	}
	if this.max_settle_volume == this.MAX {
		return supply + this.settle_volume.value
	}
	volume := supply.value + this.settle_volume.value
	volume *= this.max_settle_volume
	volume /= this.MAX
	return volume
}

func (this *AssetDataStructure) update_median_feeds(ctime SecondType) int {
	this.cfeed_publication_time = ctime
	var cfeeds []PriceType

	for f := range this.feeds {
		if (ctime-f.second.first).to_seconds() < this.feed_lifetime_sec && f.second.first != 0 {
			cfeed.emplace_back(f.second.second)
			this.cfeed_publication_time = min(this.cfeed_publication_time, f.second.first)
		}
	}

	if len(cfeed) < this.minimum_feeds {
		this.cfeed_publication_time = ctime
		cfeed = make(PriceType)
		return 0
	}

	if len(cfeed) == 1 {
		cfeed = move(cfeed.front())
		return 0
	}

	median := cfeeds[len(cfeeds)/2]
	cfeed = median

	return 0

}
