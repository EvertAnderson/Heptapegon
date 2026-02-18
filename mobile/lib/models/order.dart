import 'package:freezed_annotation/freezed_annotation.dart';

part 'order.freezed.dart';
part 'order.g.dart';

enum OrderStatus {
  @JsonValue('pending') pending,
  @JsonValue('paid') paid,
  @JsonValue('ready') ready,
  @JsonValue('completed') completed,
  @JsonValue('cancelled') cancelled,
}

@freezed
class OrderItem with _$OrderItem {
  const factory OrderItem({
    required String id,
    @JsonKey(name: 'order_id') required String orderId,
    @JsonKey(name: 'product_name') required String productName,
    required int quantity,
    @JsonKey(name: 'unit_price') required double unitPrice,
  }) = _OrderItem;

  factory OrderItem.fromJson(Map<String, dynamic> json) =>
      _$OrderItemFromJson(json);
}

@freezed
class Order with _$Order {
  const factory Order({
    required String id,
    @JsonKey(name: 'customer_id') required String customerId,
    @JsonKey(name: 'business_id') required String businessId,
    required List<OrderItem> items,
    @JsonKey(name: 'total_amount') required double totalAmount,
    required OrderStatus status,
    // Only present immediately after successful payment
    String? pin,
    @JsonKey(name: 'stripe_payment_id') String? stripePaymentId,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    @JsonKey(name: 'updated_at') required DateTime updatedAt,
  }) = _Order;

  factory Order.fromJson(Map<String, dynamic> json) => _$OrderFromJson(json);
}

// ─── Request models ──────────────────────────────────────────────────────────

@freezed
class CreateOrderItemRequest with _$CreateOrderItemRequest {
  const factory CreateOrderItemRequest({
    @JsonKey(name: 'product_name') required String productName,
    required int quantity,
    @JsonKey(name: 'unit_price') required double unitPrice,
  }) = _CreateOrderItemRequest;

  factory CreateOrderItemRequest.fromJson(Map<String, dynamic> json) =>
      _$CreateOrderItemRequestFromJson(json);
}

@freezed
class CreateOrderRequest with _$CreateOrderRequest {
  const factory CreateOrderRequest({
    @JsonKey(name: 'business_id') required String businessId,
    required List<CreateOrderItemRequest> items,
  }) = _CreateOrderRequest;

  factory CreateOrderRequest.fromJson(Map<String, dynamic> json) =>
      _$CreateOrderRequestFromJson(json);
}
