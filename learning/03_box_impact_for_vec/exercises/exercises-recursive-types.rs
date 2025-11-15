// Exercise 4: Recursive Types
// Copy this file to: src/bin/recursive_types.rs
// Run with: cargo run --bin recursive_types

fn main() {
    println!("=== Recursive Types Require Box ===\n");

    // This is a binary tree node
    #[allow(dead_code)]
    struct TreeNode {
        value: i32,
        left: Option<Box<TreeNode>>,
        right: Option<Box<TreeNode>>,
    }

    // Build a small tree:
    //       1
    //      / \
    //     2   3
    let tree = TreeNode {
        value: 1,
        left: Some(Box::new(TreeNode {
            value: 2,
            left: None,
            right: None,
        })),
        right: Some(Box::new(TreeNode {
            value: 3,
            left: None,
            right: None,
        })),
    };

    println!("Created a tree:");
    println!("       {}", tree.value);
    println!("      / \\");
    if let Some(left) = &tree.left {
        if let Some(right) = &tree.right {
            println!("     {}   {}", left.value, right.value);
        }
    }

    println!("\nTreeNode size: {} bytes", std::mem::size_of::<TreeNode>());
    println!("Why is it fixed? Because Option<Box<TreeNode>> is fixed size!");
    println!("Without Box, TreeNode would contain TreeNode infinitely!");
}
