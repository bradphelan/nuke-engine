#pragma once
#include <mathcore/vec3.h>
namespace mathcore {
struct Mat4 {
    float m[4][4];
    Mat4();
    static Mat4 translate(const Vec3& v);
    static Mat4 scale(const Vec3& v);
    Mat4 operator*(const Mat4& o) const;
    Vec3 transform_point(const Vec3& p) const;
};
}
