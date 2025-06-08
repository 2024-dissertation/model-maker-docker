#!/usr/bin/env python
#! -*- encoding: utf-8 -*-

import bpy
import sys
import os
from math import radians
from mathutils import Vector

print("Blender version:", bpy.app.version_string)
print("PBYTHON version:", bpy.app)


def convert_obj_to_glb(input_path, output_path):
    # Clear existing mesh objects
    bpy.ops.object.select_all(action="SELECT")
    bpy.ops.object.delete()

    # Import obj file with textures
    bpy.ops.wm.obj_import(
        filepath=input_path,
    )

    for obj in bpy.context.scene.objects:
        obj.rotation_euler[1] += radians(90)  # 90 degrees in radians

    bbox_corners = [obj.matrix_world @ Vector(corner) for corner in obj.bound_box]
    center = sum(bbox_corners, Vector()) / 8
    obj.location -= center

    # Select all objects
    bpy.ops.object.select_all(action="SELECT")

    # Apply Decimate modifier to reduce polygon count
    count = 0
    for obj in bpy.context.scene.objects:
        if count > 25:
            break
        count += 1
        if obj.type == "MESH":
            target_vertex_count = 50000
            current_vertex_count = len(obj.data.vertices)

            if current_vertex_count > target_vertex_count:
                modifier = obj.modifiers.new(name="Decimate", type="DECIMATE")
                modifier.ratio = target_vertex_count / current_vertex_count
                bpy.context.view_layer.objects.active = obj
                bpy.ops.object.modifier_apply(modifier="Decimate")

    # Export as GLB with textures
    bpy.ops.export_scene.gltf(
        filepath=output_path,
        export_format="GLB",
        use_selection=True,
        export_draco_mesh_compression_enable=True,  # Enable Draco compression
        export_materials="EXPORT",  # Export materials
        export_texcoords=True,  # Export texture coordinates
        export_normals=True,  # Export normals
        export_tangents=True,  # Export tangents for normal maps
        export_apply=True,  # Apply modifiers
        export_image_format="AUTO",  # Let Blender choose the best format
        export_texture_dir="",  # Embed textures in GLB
    )


if __name__ == "__main__":
    if len(sys.argv) < 5:
        print("Usage: blender -b -P convert_obj_to_glb.py -- input.obj output.glb")
        sys.exit(1)

    input_file = sys.argv[-2]
    output_file = sys.argv[-1]

    if not os.path.exists(input_file):
        print(f"Error: Input file {input_file} does not exist")
        sys.exit(1)

    try:
        convert_obj_to_glb(input_file, output_file)
        print(f"Successfully converted {input_file} to {output_file}")
    except Exception as e:
        print(f"Error during conversion: {str(e)}")
        sys.exit(1)

