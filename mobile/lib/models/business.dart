import 'package:freezed_annotation/freezed_annotation.dart';

part 'business.freezed.dart';
part 'business.g.dart';

@freezed
class Business with _$Business {
  const factory Business({
    required String id,
    @JsonKey(name: 'owner_id') required String ownerId,
    required String name,
    required String description,
    required String address,
    required double latitude,
    required double longitude,
    required String category,
    @JsonKey(name: 'is_active') @Default(true) bool isActive,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    // Present only in nearby responses
    @JsonKey(name: 'distance_km') double? distanceKm,
  }) = _Business;

  factory Business.fromJson(Map<String, dynamic> json) =>
      _$BusinessFromJson(json);
}
