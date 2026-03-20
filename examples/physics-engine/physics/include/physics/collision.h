#pragma once
#include <mathcore/vec3.h>
namespace physics {
struct AABB {
    mathcore::Vec3 min, max;
    bool intersects(const AABB& other) const;
};
}
