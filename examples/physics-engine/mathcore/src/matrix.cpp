#include <mathcore/matrix.h>
#include <cstring>

namespace mathcore {

Mat4::Mat4() {
    std::memset(m, 0, sizeof(m));
    m[0][0] = m[1][1] = m[2][2] = m[3][3] = 1.0f;
}

Mat4 Mat4::translate(const Vec3& v) {
    Mat4 r;
    r.m[0][3] = v.x;
    r.m[1][3] = v.y;
    r.m[2][3] = v.z;
    return r;
}

Mat4 Mat4::scale(const Vec3& v) {
    Mat4 r;
    r.m[0][0] = v.x;
    r.m[1][1] = v.y;
    r.m[2][2] = v.z;
    return r;
}

Mat4 Mat4::operator*(const Mat4& o) const {
    Mat4 r;
    std::memset(r.m, 0, sizeof(r.m));
    for (int i = 0; i < 4; i++)
        for (int j = 0; j < 4; j++)
            for (int k = 0; k < 4; k++)
                r.m[i][j] += m[i][k] * o.m[k][j];
    return r;
}

Vec3 Mat4::transform_point(const Vec3& p) const {
    return Vec3(
        m[0][0] * p.x + m[0][1] * p.y + m[0][2] * p.z + m[0][3],
        m[1][0] * p.x + m[1][1] * p.y + m[1][2] * p.z + m[1][3],
        m[2][0] * p.x + m[2][1] * p.y + m[2][2] * p.z + m[2][3]
    );
}

}
