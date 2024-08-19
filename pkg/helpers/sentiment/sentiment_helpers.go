package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/hjoshi123/fintel/infra/constants"
	"github.com/hjoshi123/fintel/infra/util"
	datastore "github.com/hjoshi123/fintel/pkg/datastore/db"
	datastoreIface "github.com/hjoshi123/fintel/pkg/datastore/interface"
	"github.com/hjoshi123/fintel/pkg/models"
	infoModels "github.com/hjoshi123/fintel/pkg/models/info"
)

type SentimentHelpers struct {
	SentimentStore  datastoreIface.StockSentimentStore
	TopContentStore datastoreIface.TopContentStore
}

func NewSentimentHelpers() *SentimentHelpers {
	return &SentimentHelpers{
		SentimentStore:  datastore.NewStockSentimentStore(),
		TopContentStore: datastore.NewTopContentStore(),
	}
}

func (s *SentimentHelpers) GetSentimentForStock(ctx context.Context, ticker string, timeStart, timeEnd time.Time) (*models.StockResponse, error) {
	stockSentiments, err := s.SentimentStore.GetByTickerAndTime(ctx, ticker, timeStart, timeEnd)
	if err != nil {
		util.Log.Error().Err(err).Msg("error getting associated stock sentiment")
		return nil, err
	}

	topContent, err := s.TopContentStore.GetByTickerAndTime(ctx, ticker, timeStart, timeEnd)
	if err != nil {
		util.Log.Error().Err(err).Msg("error getting top content")
		return nil, err
	}

	stockResponse := new(models.StockResponse)
	stockResponse.Ticker = ticker

	stockResponse.NewsSentiment = make([]*models.Sentiment, 0)
	stockResponse.SocialSentiment = make([]*models.Sentiment, 0)

	newsCorrelationPrices := make([]float64, 0)
	socialCorrelationPrices := make([]float64, 0)

	newsCorrelationSenitment := make([]float64, 0)
	socialCorrelationSentiment := make([]float64, 0)

	for _, stockSentiment := range stockSentiments {
		sent := new(models.Sentiment)
		sent.DailyICI = stockSentiment.DailyIci
		sent.ID = stockSentiment.ID
		sent.Date = stockSentiment.CreatedAt.Time
		sent.Volume = stockSentiment.Chatter

		stockSentInfo := new(infoModels.StockSentimentInfo)
		if err := json.NewDecoder(bytes.NewReader(stockSentiment.Info.JSON)).Decode(stockSentInfo); err != nil {
			util.Log.Error().Err(err).Msg("error decoding stock sentiment info")
		}

		sent.PositiveCount = stockSentInfo.PositiveCount
		sent.NegativeCount = stockSentInfo.NegativeCount
		sent.NeutralCount = stockSentInfo.NeutralCount

		switch stockSentiment.R.GetSource().Name {
		case constants.StockNewsSource:
			newsCorrelationPrices = append(newsCorrelationPrices, stockSentiment.Price)
			newsCorrelationSenitment = append(newsCorrelationSenitment, stockSentiment.DailyIci)
			stockResponse.NewsSentiment = append(stockResponse.NewsSentiment, sent)
		case constants.StockSocialSource:
			socialCorrelationPrices = append(socialCorrelationPrices, stockSentiment.Price)
			socialCorrelationSentiment = append(socialCorrelationSentiment, stockSentiment.DailyIci)
			stockResponse.SocialSentiment = append(stockResponse.SocialSentiment, sent)
		}
	}

	stockResponse.NewsCorrelation, err = util.Pearson(newsCorrelationPrices, newsCorrelationSenitment)
	if err != nil {
		util.Log.Error().Err(err).Msg("error calculating news correlation")
	}

	stockResponse.SocialCorrelation, err = util.Pearson(socialCorrelationPrices, socialCorrelationSentiment)
	if err != nil {
		util.Log.Error().Err(err).Msg("error calculating social correlation")
	}

	stockResponse.TopContentsNews = make([]*models.TopContentResponse, 0)
	stockResponse.TopContentsSocial = make([]*models.TopContentResponse, 0)
	for _, content := range topContent {
		topContentResponse := new(models.TopContentResponse)
		topContentResponse.ID = content.ID
		topContentResponse.URL = content.URL
		topContentResponse.PostedDate = content.CreatedAt

		topContentInfo := new(infoModels.TopContentInfo)
		err = topContentInfo.Value(content.Info.JSON)
		if err != nil {
			util.Log.Error().Err(err).Msg("error decoding top content info")
		} else {
			topContentResponse.Title = topContentInfo.Title
			topContentResponse.Summary = topContentInfo.Summary
		}

		switch content.R.GetSource().Name {
		case constants.StockNewsSource:
			stockResponse.TopContentsNews = append(stockResponse.TopContentsNews, topContentResponse)
		case constants.StockSocialSource:
			stockResponse.TopContentsSocial = append(stockResponse.TopContentsSocial, topContentResponse)
		}
	}

	sort.Slice(stockResponse.TopContentsSocial, func(i, j int) bool {
		return stockResponse.TopContentsSocial[i].PostedDate.After(stockResponse.TopContentsSocial[j].PostedDate)
	})

	if len(stockResponse.TopContentsSocial) > 10 {
		stockResponse.TopContentsSocial = stockResponse.TopContentsSocial[:10]
	}

	return stockResponse, nil
}
