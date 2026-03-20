#pragma once
namespace mathcore {
struct Vec3 {
    float x, y, z;
    Vec3() : x(0), y(0), z(0) {}
    Vec3(float x, float y, float z) : x(x), y(y), z(z) {}
    Vec3 operator+(const Vec3& o) const;
    Vec3 operator-(const Vec3& o) const;
    Vec3 operator*(float s) const;
    float dot(const Vec3& o) const;
    Vec3 cross(const Vec3& o) const;
    float length() const;
    Vec3 normalized() const;
};
}
