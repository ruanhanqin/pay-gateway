package service

import (
	"flag"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	"github.com/pjoc-team/pay-gateway/pkg/constant"
	"github.com/pjoc-team/pay-gateway/pkg/dbservice/model"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"github.com/prometheus/common/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type PayDatabaseService struct {
	*gorm.DB
}

func (s *PayDatabaseService) FindPayNoticeLessThenTime(ctx context.Context, payNotice *pb.PayNotice) (response *pb.PayNoticeResponse, err error) {
	log := logger.ContextLog(ctx)
	notice := &model.Notmodelice{}
	if err = copier.Copy(notice, payNotice); err != nil {
		log.Errorf("Failed to copy object! error: %s", err)
		return
	}
	results := make([]model.Notice, 0)
	if results := s.Where("length(next_notify_time) > 0 and next_notify_time <= ? and status != ?", notice.NextNotifyTime, constant.NOTIFY_SUCCESS).Find(&results); results.RecordNotFound() {
		log.Errorf("Find error: %v", s.Error.Error())
		return
	}
	response = &pb.PayNoticeResponse{}
	response.PayNotices = make([]*pb.PayNotice, len(results))
	for i, notice := range results {
		payNotice := &pb.PayNotice{}
		if err = copier.Copy(payNotice, notice); err != nil {
			log.Error("Copy result error! error: %v", err.Error())
		} else {
			log.Debugf("Found result: %v by query: %v", response, payNotice)
		}
		response.PayNotices[i] = payNotice
	}
	return
}

func (s *PayDatabaseService) SavePayNotice(ctx context.Context, payNotice *pb.PayNotice) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	notice := &model.Notice{}
	if err = copier.Copy(notice, payNotice); err != nil {
		log.Errorf("Failed to copy object! error: %s", err)
		return
	}
	if dbResult := s.Create(notice); dbResult.Error != nil {
		log.Errorf("Failed to save notice! notice: %v error: %s", payNotice, err.Error())
		err = dbResult.Error
		return
	}
	log.Infof("Succeed save notice: %v", payNotice)
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return
}

func (s *PayDatabaseService) UpdatePayNotice(ctx context.Context, payNotice *pb.PayNotice) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	notice := &model.Notice{}
	if err = copier.Copy(notice, payNotice); err != nil {
		log.Errorf("Failed to copy object! error: %s", err)
		return
	}
	if dbResult := s.Model(notice).Update(notice); dbResult.Error != nil {
		err = dbResult.Error
		log.Errorf("Failed to update notice! notice: %v error: %s", payNotice, err.Error())
		return
	}
	log.Infof("Succeed update notice: %v", payNotice)
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return
}

func (s *PayDatabaseService) FindPayNotice(ctx context.Context, payNotice *pb.PayNotice) (response *pb.PayNoticeResponse, err error) {
	log := logger.ContextLog(ctx)
	notice := &model.Notice{}
	if err = copier.Copy(notice, payNotice); err != nil {
		log.Errorf("Failed to copy object! error: %s", err)
		return
	}
	results := make([]model.Notice, 0)
	if results := s.Find(&results, notice); results.RecordNotFound() {
		log.Errorf("Find error: %v", s.Error.Error())
		return
	}
	response = &pb.PayNoticeResponse{}
	response.PayNotices = make([]*pb.PayNotice, len(results))
	for i, notice := range results {
		payNotice := &pb.PayNotice{}
		if err = copier.Copy(payNotice, notice); err != nil {
			log.Error("Copy result error! error: %v", err.Error())
		} else {
			log.Debugf("Found result: %v by query: %v", response, payNotice)
		}
		response.PayNotices[i] = payNotice
	}

	return
}

func (s *PayDatabaseService) SavePayNotifyOk(ctx context.Context, payNoticeOkRequest *pb.PayNoticeOk) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	tx := s.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	noticeOk := &model.NoticeOk{}
	if err = copier.Copy(noticeOk, payNoticeOkRequest); err != nil {
		log.Errorf("Failed to copy object! error: %s", err)
		tx.Rollback()
		return
	}
	if dbResult := s.Create(noticeOk); dbResult.Error != nil {
		log.Errorf("Failed to save ok order! order: %v error: %s", payNoticeOkRequest, dbResult.Error.Error())
		err = dbResult.Error
		tx.Rollback()
		return
	}
	notice := &model.Notice{GatewayOrderId: payNoticeOkRequest.GatewayOrderId}
	notice.Status = constant.ORDER_STATUS_SUCCESS
	if update := s.Model(notice).Update(notice); update.Error != nil {
		log.Errorf("Failed to update notice!")
		tx.Rollback()
		return
	}
	err = tx.Commit().Error

	log.Infof("Succeed save ok notice: %v", payNoticeOkRequest)
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return
}

func (s *PayDatabaseService) FindPayNotifyOk(ctx context.Context, payNoticeOk *pb.PayNoticeOk) (response *pb.PayNoticeOkResponse, err error) {
	log := logger.ContextLog(ctx)
	noticeOk := &model.NoticeOk{}
	if err = copier.Copy(noticeOk, payNoticeOk); err != nil {
		log.Errorf("Failed to copy object! error: %s", err)
		return
	}
	results := make([]model.NoticeOk, 0)
	if results := s.Find(&results, noticeOk); results.RecordNotFound() {
		log.Errorf("Find error: %v", s.Error.Error())
		return
	}
	response = &pb.PayNoticeOkResponse{}
	response.PayNoticeOks = make([]*pb.PayNoticeOk, len(results))

	for i, noticeOk := range results {
		payNoticeOk := &pb.PayNoticeOk{}
		if err = copier.Copy(payNoticeOk, noticeOk); err != nil {
			log.Error("Copy result error! error: %v", err.Error())
		} else {
			log.Debugf("Found result: %v by query: %v", response, payNoticeOk)
		}
		response.PayNoticeOks[i] = payNoticeOk
	}

	if err = copier.Copy(&response.PayNoticeOks, results); err != nil {
		log.Error("Copy result error! error: %v", err.Error())
	} else {
		log.Debugf("Found result: %v by query: %v", response, payNoticeOk)
	}
	return
}

func (s *PayDatabaseService) UpdatePayNoticeOk(ctx context.Context, payNoticeOk *pb.PayNoticeOk) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	noticeOk := &model.NoticeOk{}
	if err = copier.Copy(noticeOk, payNoticeOk); err != nil {
		log.Errorf("Failed to copy object! error: %s", err)
		return
	}
	if dbResult := s.Model(noticeOk).Update(noticeOk); dbResult.Error != nil {
		log.Errorf("Failed to save ok notice! noticeOk: %v error: %s", payNoticeOk, err.Error())
		err = dbResult.Error
		return
	}
	log.Infof("Succeed save ok notice: %v", payNoticeOk)
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return
}

func (s *PayDatabaseService) FindPayOrder(ctx context.Context, orderRequest *pb.PayOrder) (response *pb.PayOrderResponse, err error) {
	log := logger.ContextLog(ctx)
	order := &model.PayOrder{}
	if err = copier.Copy(order, orderRequest); err != nil {
		log.Errorf("Failed to copy object! error: %s", err)
		return
	}
	results := make([]model.PayOrder, 0)
	if results := s.Find(&results, order); results.RecordNotFound() {
		log.Errorf("Find error: %v", s.Error.Error())
		return
	}
	if log.IsDebugEnabled() {
		log.Debugf("Find order: %v by order: %v", results, orderRequest)
	}
	response = &pb.PayOrderResponse{}
	response.PayOrders = make([]*pb.PayOrder, len(results))
	for i, payOrder := range results {
		order := &pb.PayOrder{}
		order.BasePayOrder = &pb.BasePayOrder{}
		response.PayOrders[i] = order

		if err = copier.Copy(response.PayOrders[i], payOrder); err != nil {
			log.Error("Copy result error! error: %v", err.Error())
		} else if err = copier.Copy(order.BasePayOrder, payOrder); err != nil {
			log.Error("Copy result error! error: %v", err.Error())
		} else {
			log.Debugf("Found result: %v by query: %v", response, orderRequest)
		}
	}

	return
}

func (s *PayDatabaseService) FindPayOrderOk(ctx context.Context, orderOkRequest *pb.PayOrderOk) (response *pb.PayOrderOkResponse, err error) {
	log := logger.ContextLog(ctx)
	orderOk := &model.PayOrderOk{}
	if err = copier.Copy(orderOk, orderOkRequest); err != nil {
		log.Errorf("Failed to copy object! error: %s", err)
		return
	}
	results := make([]model.PayOrderOk, 0)
	if results := s.Find(&results, orderOk); results.RecordNotFound() {
		log.Errorf("Find error: %v", s.Error.Error())
		return
	}
	response = &pb.PayOrderOkResponse{}
	response.PayOrderOks = make([]*pb.PayOrderOk, len(results))
	for i, payOrderOk := range results {
		orderOk := &pb.PayOrderOk{}
		orderOk.BasePayOrder = &pb.BasePayOrder{}
		response.PayOrderOks[i] = orderOk

		if err = copier.Copy(orderOk, payOrderOk); err != nil {
			log.Error("Copy result error! error: %v", err.Error())
		} else if err = copier.Copy(orderOk.BasePayOrder, payOrderOk); err != nil {
			log.Error("Copy result error! error: %v", err.Error())
		} else {
			log.Debugf("Found result: %v by query: %v", response, orderOkRequest)
		}
	}
	return
}

func (s *PayDatabaseService) SavePayOrder(ctx context.Context, orderRequest *pb.PayOrder) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	order := &model.PayOrder{}
	if err = copier.Copy(order, orderRequest); err != nil {
		log.Errorf("Failed to copy object! error: %s", err)
		return
	}
	if dbResult := s.Create(order); dbResult.Error != nil {
		log.Errorf("Failed to save order! order: %v error: %s", orderRequest, dbResult.Error.Error())
		err = dbResult.Error
		return
	}
	log.Infof("Succeed save order: %v", orderRequest)
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return
}

func (s *PayDatabaseService) UpdatePayOrder(ctx context.Context, orderRequest *pb.PayOrder) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	order := &model.PayOrder{}
	if err = copier.Copy(order, orderRequest); err != nil {
		log.Errorf("Failed to copy object! error: %s", err)
		return
	}
	if dbResult := s.Model(order).Update(order); dbResult.Error != nil {
		log.Errorf("Failed to update order! order: %v error: %s", orderRequest, dbResult.Error.Error())
		err = dbResult.Error
		return
	}
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	log.Infof("Succeed update order: %v result: %v", orderRequest, result)
	return
}

func (s *PayDatabaseService) SavePayOrderOk(ctx context.Context, orderOkRequest *pb.PayOrderOk) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	tx := s.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	order := &model.PayOrderOk{}
	if err = copier.Copy(order, orderOkRequest); err != nil {
		log.Errorf("Failed to copy object! error: %s", err)
		tx.Rollback()
		return
	}
	if dbResult := s.Create(order); dbResult.Error != nil {
		log.Errorf("Failed to save ok order! order: %v error: %s", orderOkRequest, dbResult.Error.Error())
		err = dbResult.Error
		tx.Rollback()
		return
	}
	payOrder := &model.PayOrder{BasePayOrder: model.BasePayOrder{GatewayOrderId: orderOkRequest.BasePayOrder.GatewayOrderId}}
	payOrder.OrderStatus = constant.ORDER_STATUS_SUCCESS
	if update := s.Model(payOrder).Update(payOrder); update.Error != nil {
		log.Errorf("Failed to update order!")
		tx.Rollback()
		return
	}
	err = tx.Commit().Error

	log.Infof("Succeed save ok order: %v", orderOkRequest)
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return
}

func (s *PayDatabaseService) UpdatePayOrderOk(ctx context.Context, orderOkRequest *pb.PayOrderOk) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	order := &model.PayOrderOk{}
	if err = copier.Copy(order, orderOkRequest); err != nil {
		log.Errorf("Failed to copy object! error: %s", err)
		return
	}
	if dbResult := s.Model(order).Update(order); dbResult.Error != nil {
		log.Errorf("Failed to save ok order! order: %v error: %s", orderOkRequest, dbResult.Error.Error())
		err = dbResult.Error
		return
	}
	log.Infof("Succeed save ok order: %v", orderOkRequest)
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return
}

func (svc *PayDatabaseService) RegisterGrpc(gs *grpc.Server) {
	pb.RegisterPayDatabaseServiceServer(gs, svc)
}

func Init(service *service.Service, db *gorm.DB) (pb.PayDatabaseServiceServer, error) {
	svc := &PayDatabaseService{}
	svc.DB = db
	svc.Service = service
	flag.Parse()

	svc.StartGrpc(svc.RegisterGrpc)
}