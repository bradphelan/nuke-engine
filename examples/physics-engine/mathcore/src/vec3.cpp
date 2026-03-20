#include <mathcore/vec3.h>
#include "internal/simd_helpers.h"
#include <cmath>

namespace mathcore {

Vec3 Vec3::operator+(const Vec3& o) const { return Vec3(x + o.x, y + o.y, z + o.z); }
Vec3 Vec3::operator-(const Vec3& o) const { return Vec3(x - o.x, y - o.y, z - o.z); }
Vec3 Vec3::operator*(float s) const { return Vec3(x * s, y * s, z * s); }
float Vec3::dot(const Vec3& o) const { return x * o.x + y * o.y + z * o.z; }

Vec3 Vec3::cross(const Vec3& o) const {
    return Vec3(
        y * o.z - z * o.y,
        z * o.x - x * o.z,
        x * o.y - y * o.x
    );
}

float Vec3::length() const { return std::sqrt(x * x + y * y + z * z); }

Vec3 Vec3::normalized() const {
    float len2 = x * x + y * y + z * z;
    if (len2 == 0.0f) return Vec3();
    float inv = fast_rsqrt(len2);
    return Vec3(x * inv, y * inv, z * inv);
}

}
