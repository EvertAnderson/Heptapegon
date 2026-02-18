import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/business.dart';
import '../models/order.dart';

// ─── Configuration ────────────────────────────────────────────────────────────

// Override this with an env var or flavour config in production.
const String _baseUrl = 'http://localhost:8080/api/v1';

// ─── Auth token state ─────────────────────────────────────────────────────────

final authTokenProvider = StateProvider<String?>((ref) => null);

// ─── Dio instance ─────────────────────────────────────────────────────────────

final dioProvider = Provider<Dio>((ref) {
  final dio = Dio(
    BaseOptions(
      baseUrl: _baseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 15),
      headers: {'Content-Type': 'application/json'},
    ),
  );

  dio.interceptors.add(
    InterceptorsWrapper(
      onRequest: (options, handler) {
        final token = ref.read(authTokenProvider);
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
      onError: (DioException error, handler) {
        if (error.response?.statusCode == 401) {
          // Token expired — clear state; UI layer should redirect to login.
          ref.read(authTokenProvider.notifier).state = null;
        }
        handler.next(error);
      },
    ),
  );

  return dio;
});

// ─── Business API ─────────────────────────────────────────────────────────────

final businessApiProvider = Provider<BusinessApi>(
  (ref) => BusinessApi(ref.watch(dioProvider)),
);

class BusinessApi {
  const BusinessApi(this._dio);
  final Dio _dio;

  /// Returns businesses within [radiusKm] km of ([lat], [lng]).
  Future<List<Business>> getNearby({
    required double lat,
    required double lng,
    double radiusKm = 5.0,
    String? category,
  }) async {
    final response = await _dio.get<Map<String, dynamic>>(
      '/businesses/nearby',
      queryParameters: {
        'lat': lat,
        'lng': lng,
        'radius': radiusKm,
        if (category != null) 'category': category,
      },
    );
    final data = response.data?['data'] as List<dynamic>? ?? [];
    return data
        .cast<Map<String, dynamic>>()
        .map(Business.fromJson)
        .toList();
  }

  Future<Business> getById(String id) async {
    final response =
        await _dio.get<Map<String, dynamic>>('/businesses/$id');
    return Business.fromJson(response.data!);
  }
}

// ─── Order API ────────────────────────────────────────────────────────────────

final orderApiProvider = Provider<OrderApi>(
  (ref) => OrderApi(ref.watch(dioProvider)),
);

class OrderApi {
  const OrderApi(this._dio);
  final Dio _dio;

  /// Creates an order and returns the response that includes the one-time PIN.
  Future<Order> createOrder(CreateOrderRequest request) async {
    final response = await _dio.post<Map<String, dynamic>>(
      '/orders',
      data: request.toJson(),
    );
    return Order.fromJson(response.data!);
  }

  Future<Order> getOrder(String id) async {
    final response =
        await _dio.get<Map<String, dynamic>>('/orders/$id');
    return Order.fromJson(response.data!);
  }

  Future<List<Order>> listOrders() async {
    final response = await _dio.get<Map<String, dynamic>>('/orders');
    final data = response.data?['data'] as List<dynamic>? ?? [];
    return data.cast<Map<String, dynamic>>().map(Order.fromJson).toList();
  }

  /// Called by the business app to validate the customer's PIN at pickup.
  Future<void> validatePin(String orderId, String pin) async {
    await _dio.post<void>(
      '/orders/$orderId/validate-pin',
      data: {'pin': pin},
    );
  }
}

// ─── Auth API ─────────────────────────────────────────────────────────────────

final authApiProvider = Provider<AuthApi>(
  (ref) => AuthApi(ref.watch(dioProvider), ref),
);

class AuthApi {
  const AuthApi(this._dio, this._ref);
  final Dio _dio;
  final Ref _ref;

  Future<String> login(String email, String password) async {
    // Auth endpoints are on /auth, not /api/v1, so strip the baseUrl prefix.
    final response = await _dio.post<Map<String, dynamic>>(
      // ignore: avoid_redundant_argument_values
      '/../../auth/login',
      data: {'email': email, 'password': password},
    );
    final token = response.data!['token'] as String;
    _ref.read(authTokenProvider.notifier).state = token;
    return token;
  }

  Future<String> register({
    required String name,
    required String email,
    required String password,
    required String role,
  }) async {
    final response = await _dio.post<Map<String, dynamic>>(
      '/../../auth/register',
      data: {
        'name': name,
        'email': email,
        'password': password,
        'role': role,
      },
    );
    final token = response.data!['token'] as String;
    _ref.read(authTokenProvider.notifier).state = token;
    return token;
  }
}
